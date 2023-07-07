


编译脚本：

```bash

GOOS=js GOARCH=wasm go build -o ./target/main.wasm   main.go 

or
tinygo build -o ./target/main.wasm -target wasm ./main.go

tinygo build -o ./target/main.wasm -target wasi ./main.go

 
cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js ./target/wasm_exec_tiny.js

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./target 

```

