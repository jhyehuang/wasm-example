/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wavm

import (
	"fmt"
	"github.com/jhyehuang/wasm-example/src/wavm/common"
	"io/ioutil"
	"testing"

	vmPb "chainmaker.org/chainmaker/pb-go/v2/vm"

	logger2 "chainmaker.org/chainmaker/logger/v2"
	accessPb "chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
)

const (
	ContractName    = "ContractTest001"
	ContractVersion = "1.0.0"
	ChainId         = "chain02"
	BlockVersion    = uint32(1)
)

func readWasmFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func prepareContract(filepath string, t *testing.T) ([]byte, common.Contract, *logger2.CMLogger) {
	wasmBytes, err := readWasmFile(filepath)
	if err != nil {
		t.Fatalf("read wasm file error: %v", err)
	}

	contractId := common.Contract{
		Name:    ContractName,
		Version: ContractVersion,
	}

	logger := logger2.GetLogger("unit_test")

	return wasmBytes, contractId, logger
}

func fillingBaseParams(parameters map[string][]byte) {
	parameters[protocol.ContractTxIdParam] = []byte("TX_ID")
	parameters[protocol.ContractCreatorOrgIdParam] = []byte("CREATOR_ORG_ID")
	parameters[protocol.ContractCreatorRoleParam] = []byte("CREATOR_ROLE")
	parameters[protocol.ContractCreatorPkParam] = []byte("CREATOR_PK")
	parameters[protocol.ContractSenderOrgIdParam] = []byte("SENDER_ORG_ID")
	parameters[protocol.ContractSenderRoleParam] = []byte("SENDER_ROLE")
	parameters[protocol.ContractSenderPkParam] = []byte("SENDER_PK")
	parameters[protocol.ContractBlockHeightParam] = []byte("111")
}

type SnapshotMock struct {
	cache map[string][]byte
}

func (s SnapshotMock) GetBlockFingerprint() string {
	panic("implement me")
}

func (s SnapshotMock) GetKeys(txExecSeq int, keys []*vmPb.BatchKey) ([]*vmPb.BatchKey, error) {
	panic("implement me")
}

func (s SnapshotMock) BuildDAG(isSql bool, txRWSetTable []*commonPb.TxRWSet) *commonPb.DAG {
	panic("implement me")
}

func (s SnapshotMock) ApplyBlock(block *commonPb.Block, txRWSetMap map[string]*commonPb.TxRWSet) {
	panic("implement me")
}

func (s SnapshotMock) GetSpecialTxTable() []*commonPb.Transaction {
	panic("implement me")
}

func (s SnapshotMock) GetBlockTimestamp() int64 {
	panic("implement me")
}

func (s SnapshotMock) ApplyTxSimContext(context protocol.TxSimContext, txType protocol.ExecOrderTxType,
	b bool, b2 bool) (bool, int) {
	panic("implement me")
}

func (s SnapshotMock) GetBlockchainStore() protocol.BlockchainStore {
	panic("implement me")
}

func (s SnapshotMock) GetKey(txExecSeq int, contractName string, key []byte) ([]byte, error) {
	combinedKey := fmt.Sprintf("%d-%s-%s", txExecSeq, contractName, key)
	return s.cache[combinedKey], nil
}

func (s SnapshotMock) GetTxRWSetTable() []*commonPb.TxRWSet {
	panic("implement me")
}

func (s SnapshotMock) GetTxResultMap() map[string]*commonPb.Result {
	panic("implement me")
}

func (s SnapshotMock) GetSnapshotSize() int {
	return 0
}

func (s SnapshotMock) GetTxTable() []*commonPb.Transaction {
	panic("implement me")
}

func (s SnapshotMock) GetPreSnapshot() protocol.Snapshot {
	panic("implement me")
}

func (s SnapshotMock) SetPreSnapshot(snapshot protocol.Snapshot) {
	panic("implement me")
}

func (s SnapshotMock) GetBlockHeight() uint64 {
	panic("implement me")
}

func (s SnapshotMock) GetBlockProposer() *accessPb.Member {
	panic("implement me")
}

func (s SnapshotMock) IsSealed() bool {
	panic("implement me")
}

func (s SnapshotMock) Seal() {
	panic("implement me")
}
