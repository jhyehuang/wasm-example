package wavm

import (
	"chainmaker.org/chainmaker/logger/v2"
	"chainmaker.org/chainmaker/protocol/v2"
	"fmt"
	"github.com/jhyehuang/wasm-example/pkg/utils"
	wasmergo "github.com/jhyehuang/wasm-example/pkg/wasmer-go"
	"github.com/jhyehuang/wasm-example/src/wavm/common"
)

// wrappedInstance wraps instance with id and other info
type wrappedInstance struct {
	// id
	id string
	// wasmergo instance provided by wasmer
	wasmInstance *wasmergo.Instance
	// lastUseTime, unix timestamp in ms
	lastUseTime int64
	// createTime, unix timestamp in ms
	createTime int64
	// errCount, current instance invoke method error count
	errCount int32
}

// vmPool, each contract has a vm pool providing multiple vm instances to call
// vm pool can grow and shrink on demand
type vmPool struct {
	// the corresponding contract info
	contractId *common.Contract
	byteCode   []byte
	store      *wasmergo.Store
	module     *wasmergo.Module
	// wasmergo instance pool
	instances chan *wrappedInstance
	// current instance size in pool
	currentSize int32
	// use count from last refresh
	useCount int32
	// total delay (in ms) from last refresh
	totalDelay int32
	// total application count for pool grow
	// if we cannot get instance right now, applyGrowCount++
	applyGrowCount int32
	// apply signal channel
	applySignalC    chan struct{}
	closeC          chan struct{}
	resetC          chan struct{}
	removeInstanceC chan struct{}
	addInstanceC    chan struct{}
	log             *logger.CMLogger
}

// RuntimeInstance wasm runtime
type RuntimeInstance struct {
	pool *vmPool
	log  *logger.CMLogger
}

// Pool comment at next version
func (r *RuntimeInstance) Pool() *vmPool {
	return r.pool
}

// Invoke contract by call vm, implement protocol.RuntimeInstance
func (r *RuntimeInstance) Invoke(contract *common.Contract, method string, byteCode []byte,
	parameters map[string][]byte, gasUsed uint64) (
	contractResult *common.ContractResult) {

	startTime := utils.CurrentTimeMillisSeconds()
	logStr := fmt.Sprintf("wasmer runtime invoke[%s]: ", contract.Name)

	// set default return value
	contractResult = &common.ContractResult{
		Code:    uint32(0),
		Result:  nil,
		Message: "",
	}

	var instanceInfo *wrappedInstance
	defer func() {
		endTime := utils.CurrentTimeMillisSeconds()
		logStr := fmt.Sprintf(" used time %d", endTime-startTime)
		r.log.Debugf(logStr)
		panicErr := recover()
		if panicErr != nil {
			contractResult.Code = 1
			contractResult.Message = fmt.Sprint(panicErr)
			if instanceInfo != nil {
				instanceInfo.errCount++
			}
		}
	}()

	instanceInfo = r.pool.GetInstance()
	defer r.pool.RevertInstance(instanceInfo)

	instance := instanceInfo.wasmInstance
	instance.SetGasLimit(protocol.GasLimit - gasUsed)

	var sc = NewSimContext(method, r.log, "")
	defer sc.removeCtxPointer()
	sc.Contract = contract
	sc.ContractResult = contractResult
	sc.parameters = parameters
	sc.Instance = instance

	err := sc.CallMethod(instance)
	if err != nil {
		r.log.Errorf("contract invoke failed, %s, tx: %s", err)
	}

	// gas Log
	gas := protocol.GasLimit - instance.GetGasRemaining()
	if instance.GetGasRemaining() <= 0 {
		err = fmt.Errorf("contract invoke failed, out of gas %d/%d, tx: %s", gas, int64(protocol.GasLimit),
			"")
	}
	logStr += fmt.Sprintf("used gas %d ", gas)
	contractResult.GasUsed = gas

	if err != nil {
		contractResult.Code = 1
		msg := fmt.Sprintf("contract invoke failed, %s", err.Error())
		r.log.Errorf(msg)
		contractResult.Message = msg
		if method != "init_contract" {
			instanceInfo.errCount++
		}
		return
	}
	contractResult.GasUsed = gas
	return
}
