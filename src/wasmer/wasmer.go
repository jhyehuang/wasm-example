package main

import (
	"fmt"
	"io/ioutil"

	wasm "github.com/wasmerio/wasmer-go/wasmer"
)

func main() {
	wasmBytes, err := ioutil.ReadFile("target/main.wasm")
	check(err)
	fmt.Println("wasmBytes:", wasmBytes)

	// Create an Engine
	engine := wasm.NewEngine()

	// Create a Store
	store := wasm.NewStore(engine)

	// Let's compile the module.
	module, err := wasm.NewModule(store, wasmBytes)

	if err != nil {
		fmt.Println("Failed to compile module:", err)
	}

	// Create an empty import object.
	importObject := wasm.NewImportObject()

	// Let's instantiate the WebAssembly module.
	instance, err := wasm.NewInstance(module, importObject)
	if err != nil {
		panic(fmt.Sprintln("Failed to instantiate the module:", err))
	}

	// Now let's execute the `sum` function.
	sum, err := instance.Exports.GetFunction("CallMeFromJavascript")

	if err != nil {
		panic(fmt.Sprintln("Failed to get the `add_one` function:", err))
	}

	result, err := sum(1, 2)

	if err != nil {
		panic(fmt.Sprintln("Failed to call the `add_one` function:", err))
	}

	fmt.Println("Results of `sum`:", result)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
