/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wavm

import (
	"bytes"
	"fmt"
	"github.com/jhyehuang/wasm-example/pkg/log"
	"testing"

	"chainmaker.org/chainmaker/protocol/v2"
)

func readWriteSet(txSimContext protocol.TxSimContext) ([]byte, error) {
	rwSet := txSimContext.GetTxRWSet(true)
	fmt.Printf("rwSet = %v \n", rwSet)

	var result []byte
	for _, w := range rwSet.TxWrites {
		if bytes.Equal(w.Key, []byte("count#test_key")) {
			result = w.Value
		}
	}
	if result == nil {
		return nil, fmt.Errorf("write set contain no 'count#test_key'")
	}

	return result, nil
}

// TestInvoke comment at next version
func TestInvoke(t *testing.T) {

	wasmBytes, contractId, logger := prepareContract("./testdata/helloworld.wasm", t)

	vmPool, err := newVmPool(&contractId, wasmBytes, logger)
	if err != nil {
		t.Fatalf("create vmPool error: %v", err)
	}

	defer func() {
		vmPool.close()
	}()

	runtimeInst := RuntimeInstance{
		pool: vmPool,
		log:  logger,
	}

	parameters := make(map[string][]byte)
	parameters["key"] = []byte("test_key")
	fillingBaseParams(parameters)

	// 测试一次调用结果是否正确
	ret := runtimeInst.Invoke(&contractId, "increase", wasmBytes, parameters, 0)
	log.Infof("ret = %v", ret)
	// 测试第二次调用结果是否正确
	runtimeInst.Invoke(&contractId, "increase", wasmBytes, parameters, 0)
	log.Infof("ret = %v", ret)

}
