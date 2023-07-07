package wavm

import (
	wasmergo "github.com/jhyehuang/wasm-example/pkg/wasmer-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRawFunction(t *testing.T) {
	engine := wasmergo.NewEngine()
	store := wasmergo.NewStore(engine)
	wasmBytes, _, _ := prepareContract("./testdata/helloworld.wasm", t)
	module, err := wasmergo.NewModule(
		store,
		wasmBytes, nil,
	)
	assert.NoError(t, err)

	instance, err := wasmergo.NewInstance(module, wasmergo.NewImportObject())
	assert.NoError(t, err)

	sum, err := instance.Exports.GetRawFunction("callMeFromJavascript")
	assert.NoError(t, err)
	assert.Equal(t, sum.ParameterArity(), uint(2))
	assert.Equal(t, sum.ResultArity(), uint(1))

	result, err := sum.Call(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, result, int32(3))
}
