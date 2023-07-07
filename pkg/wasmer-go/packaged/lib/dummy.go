// See https://github.com/golang/go/issues/26366.
package lib

import (
	_ "github.com/jhyehuang/wasm-example/pkg/wasmer-go/packaged/lib/darwin-aarch64"
	_ "github.com/jhyehuang/wasm-example/pkg/wasmer-go/packaged/lib/darwin-amd64"
	_ "github.com/jhyehuang/wasm-example/pkg/wasmer-go/packaged/lib/linux-aarch64"
	_ "github.com/jhyehuang/wasm-example/pkg/wasmer-go/packaged/lib/linux-amd64"
	_ "github.com/jhyehuang/wasm-example/pkg/wasmer-go/packaged/lib/windows-amd64"
)
