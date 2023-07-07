package wavm

import (
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	"chainmaker.org/chainmaker/logger/v2"
	"fmt"
	"github.com/jhyehuang/wasm-example/pkg/log"
	"github.com/jhyehuang/wasm-example/pkg/utils"
	wasmergo "github.com/jhyehuang/wasm-example/pkg/wasmer-go"
	"github.com/jhyehuang/wasm-example/src/wavm/common"
	"sync/atomic"
	"time"
)

const (
	// refresh vmPool time, use for grow or shrink
	defaultRefreshTime = time.Hour * 12
	// the max pool size for every contract
	defaultMaxSize = 50
	// the min pool size
	defaultMinSize = 5
	// grow pool size
	defaultChangeSize = 5
	// if get instance avg time greater than this value, should grow pool, Millisecond as unit
	defaultDelayTolerance = 10
	// if apply times greater than this value, should grow pool
	defaultApplyThreshold = 100
	// if wasmer instance invoke error more than N times, should close and discard this instance
	defaultDiscardCount = 10
)

// GetInstance get a vm instance to run contract
// should be followed by defer resetInstance
func (p *vmPool) GetInstance() *wrappedInstance {

	var instance *wrappedInstance
	// get instance from vm pool
	select {
	case instance = <-p.instances:
		// concurrency safe here
		atomic.AddInt32(&p.useCount, 1)
		instance.lastUseTime = utils.CurrentTimeMillisSeconds()
		return instance
	default:
		// nothing
	}
	if instance == nil {
		log.Debugf("can't get wrappedInstance from vmPool.")
	}

	// if we cannot get it right now, send apply signal and wait
	// add wait time to total delay
	curTimeMS1 := utils.CurrentTimeMillisSeconds()
	go func() {
		p.applySignalC <- struct{}{}
		log.Debugf("send 'applySignal' to vmPool.")
	}()

	instance = <-p.instances
	log.Debugf("got an wrappedInstance from vmPool.")
	atomic.AddInt32(&p.useCount, 1)
	curTimeMS2 := utils.CurrentTimeMillisSeconds()
	instance.lastUseTime = curTimeMS2
	elapsedTimeMS := int32(curTimeMS2 - curTimeMS1)
	atomic.AddInt32(&p.totalDelay, elapsedTimeMS)

	return instance
}

// RevertInstance revert instance to pool
func (p *vmPool) RevertInstance(instance *wrappedInstance) {
	if p.shouldDiscard(instance) {
		go func() {
			p.removeInstanceC <- struct{}{}
			p.addInstanceC <- struct{}{}
			p.CloseInstance(instance)
		}()
	} else {
		p.instances <- instance
	}
}

// shouldDiscard discard instance when
// error count times more than defaultDiscardCount
func (p *vmPool) shouldDiscard(instance *wrappedInstance) bool {
	return instance.errCount > defaultDiscardCount
}

// CloseInstance close a wasmer instance directly, for cross contract call
func (p *vmPool) CloseInstance(instance *wrappedInstance) {
	if instance != nil {
		if err := CallDeallocate(instance.wasmInstance); err != nil {
			p.log.Errorf("CallDeallocate(...) error: %v", err)
		}
		instance.wasmInstance.Close()
		instance = nil
	}
}

// NewInstance create a wasmer instance directly, for cross contract call
func (p *vmPool) NewInstance() (*wrappedInstance, error) {
	return p.newInstanceFromModule()
}

func (p *vmPool) newInstanceFromModule() (*wrappedInstance, error) {

	imports := wasmergo.NewImportObject()

	wasmInstance, err := wasmergo.NewInstance(p.module, imports)
	if err != nil {
		p.log.Errorf("newInstanceFromModule fail: %s", err.Error())
		return nil, err
	}

	instance := &wrappedInstance{
		id:           uuid.GetUUID(),
		wasmInstance: wasmInstance,
		lastUseTime:  utils.CurrentTimeMillisSeconds(),
		createTime:   utils.CurrentTimeMillisSeconds(),
		errCount:     0,
	}
	return instance, nil
}

func newVmPool(contractId *common.Contract, byteCode []byte, log *logger.CMLogger) (*vmPool, error) {
	store := wasmergo.NewStore(wasmergo.NewUniversalEngine())
	if err := wasmergo.ValidateModule(store, byteCode); err != nil {
		return nil, fmt.Errorf("[%s_%s], byte code validation failed, err = %v", contractId.Name, contractId.Version, err)
	}

	module, err := wasmergo.NewModule(store, byteCode, log)
	if err != nil {
		return nil, fmt.Errorf("[%s_%s], byte code compile failed", contractId.Name, contractId.Version)
	}

	vmPool := &vmPool{
		contractId:      contractId,
		byteCode:        byteCode,
		store:           store,
		module:          module,
		instances:       make(chan *wrappedInstance, 5),
		currentSize:     0,
		useCount:        0,
		totalDelay:      0,
		applyGrowCount:  0,
		applySignalC:    make(chan struct{}),
		removeInstanceC: make(chan struct{}),
		addInstanceC:    make(chan struct{}),
		closeC:          make(chan struct{}),
		resetC:          make(chan struct{}),
		log:             log,
	}

	instance, err := vmPool.newInstanceFromModule()
	if err != nil {
		return nil, fmt.Errorf("[%s_%s], byte code compile failed, %s", contractId.Name, contractId.Version, err.Error())
	}

	instance.wasmInstance.Close()
	log.Infof("vm pool verify byteCode finish.")

	go vmPool.startRefreshingLoop()
	log.Infof("vm pool startRefreshingLoop...")
	return vmPool, nil
}

// startRefreshingLoop refreshing loop manages the vm pool
// all grow and shrink operations are called here
func (p *vmPool) startRefreshingLoop() {

	refreshTimer := time.NewTimer(defaultRefreshTime)
	key := p.contractId.Name + "_" + p.contractId.Version
	for {
		select {
		case <-p.applySignalC:
			log.Debug("vmPool handling an `apply` Signal")
			p.applyGrowCount++
			if p.shouldGrow() {
				log.Debugf("vmPool should grow %v wrappedInstance.", defaultChangeSize)
				p.grow(defaultChangeSize)
				p.applyGrowCount = 0
				p.log.Infof("[%s] vm pool grows by %d, the current size is %d",
					key, defaultChangeSize, p.currentSize)
			}
		case <-refreshTimer.C:
			p.log.Debugf("[%s] vmPool handling an `refresh` Signal", key)
			if p.shouldGrow() {
				p.grow(defaultChangeSize)
				p.applyGrowCount = 0
				p.log.Infof("[%s] vm pool grows by %d, the current size is %d",
					key, defaultChangeSize, p.currentSize)
			} else if p.shouldShrink() {
				p.shrink(defaultChangeSize)
				p.log.Infof("[%s] vm pool shrinks by %d, the current size is %d",
					key, defaultChangeSize, p.currentSize)
			}

			// other go routine may modify useCount & totalDelay
			// so we use atomic operation here
			atomic.StoreInt32(&p.useCount, 0)
			atomic.StoreInt32(&p.totalDelay, 0)
			refreshTimer.Reset(defaultRefreshTime)
		case <-p.closeC:
			p.log.Debugf("[%s] vmPool handling an `close` Signal", key)
			refreshTimer.Stop()
			for p.currentSize > 0 {
				instance := <-p.instances
				if err := CallDeallocate(instance.wasmInstance); err != nil {
					p.log.Errorf("CallDeallocate(...) error: %v", err)
				}
				instance.wasmInstance.Close()
				p.currentSize--
			}
			close(p.instances)
			return
		case <-p.resetC:
			p.log.Debugf("[%s] vmPool handling an `reset` Signal", key)
			for p.currentSize > 0 {
				instance := <-p.instances
				if err := CallDeallocate(instance.wasmInstance); err != nil {
					p.log.Errorf("CallDeallocate(...) error: %v", err)
				}
				instance.wasmInstance.Close()
				p.currentSize--
			}
			close(p.instances)
			p.instances = make(chan *wrappedInstance, defaultMaxSize)
			p.grow(defaultMinSize)
		case <-p.removeInstanceC:
			p.log.Debugf("[%s] vmPool handling an `remove instance` Signal", key)
			p.currentSize--
		case <-p.addInstanceC:
			p.log.Debugf("[%s] vmPool handling an `add instance` Signal", key)
			p.grow(1)
		}
	}
}

// shouldGrow grow vm pool when
// 1. current size + grow size <= max size, AND
// 2.1. apply count >= apply threshold, OR
// 2.2. average delay > delay tolerance (int operation here is safe)
func (p *vmPool) shouldGrow() bool {
	if p.currentSize < defaultMinSize {
		return true
	}

	if p.currentSize+defaultChangeSize <= defaultMaxSize {
		if p.applyGrowCount > defaultApplyThreshold {
			return true
		}

		if p.getAverageDelay() > int32(defaultDelayTolerance) {
			return true
		}

		if p.currentSize < int32(defaultMinSize) {
			return true
		}
	}
	return false
}

func (p *vmPool) grow(count int32) {
	for count > 0 {
		size := int32(defaultChangeSize)
		if count < size {
			size = count
		}
		count -= size

		for i := int32(0); i < size; i++ {
			instance, _ := p.newInstanceFromModule()
			p.instances <- instance
			atomic.AddInt32(&p.currentSize, 1)
		}
		p.log.Infof("vm pool grow size = %d", size)
	}
}

// shouldShrink shrink vm pool when
// 1. current size > min size, AND
// 2. average delay <= delay tolerance (int operation here is safe)
func (p *vmPool) shouldShrink() bool {
	if p.currentSize > defaultMinSize && p.getAverageDelay() <=
		int32(defaultDelayTolerance) && p.currentSize > defaultChangeSize {
		return true
	}
	return false
}

func (p *vmPool) shrink(count int32) {
	for i := int32(0); i < count; i++ {
		instance := <-p.instances
		if err := CallDeallocate(instance.wasmInstance); err != nil {
			p.log.Errorf("CallDeallocate(...) error: %v", err)
		}
		instance.wasmInstance.Close()
		instance = nil
		p.currentSize--
	}
}

// getAverageDelay average delay calculation here maybe not so accurate due to concurrency
// but we can still use it to decide grow/shrink or not
func (p *vmPool) getAverageDelay() int32 {
	delay := atomic.LoadInt32(&p.totalDelay)
	count := atomic.LoadInt32(&p.useCount)
	if count == 0 {
		return 0
	}
	return delay / count
}

// reset the pool instances
func (p *vmPool) reset() {
	p.resetC <- struct{}{}
}

// close the pool
func (p *vmPool) close() {
	close(p.closeC)
}
