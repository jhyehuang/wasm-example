SHELL=/usr/bin/env bash

all: memory

http-server:
	go run main.go

target/wasm_exec.js:
	cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js ./target/wasm_exec_tiny.js

target/main.wasm: main.go
	tinygo build -o ./target/main.wasm -target wasm ./main.go

memory-build:
	tinygo build -o ./target/main.wasm -target wasm ./src/memory/main.go
	cp ./src/memory/index.js ./target/index.js

memory: memory-build target/wasm_exec.js http-server

import-js-func-build:
	tinygo build -o ./target/main.wasm -target wasm ./src/import-js-func/main.go
	cp ./src/import-js-func/index.js ./target/index.js

import-js-func: import-js-func-build target/wasm_exec.js http-server

graphics-build:
	tinygo build -o ./target/main.wasm -target wasm ./src/graphics/main.go
	cp ./src/graphics/index.js ./target/index.js

graphics: graphics-build target/wasm_exec.js http-server

audio-build:
	tinygo build -o ./target/main.wasm -target wasm ./src/audio/main.go
	cp ./src/audio/index.js ./target/index.js

audio: audio-build target/wasm_exec.js http-server


wasm-build:
	tinygo build -o ./target/memory.wasm -target wasm ./src/memory/main.go
	tinygo build -o ./target/graphics.wasm -target wasm ./src/graphics/main.go
	tinygo build -o ./target/audio.wasm -target wasm ./src/audio/main.go
	tinygo build -o ./target/helloworld.wasm -target wasm ./src/hello-world/main.go

