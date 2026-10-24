#!/usr/bin/env bash
# Compila um programa COOL para WASM e executa.
# Uso: ./compilar.sh examples/minimo.cl
set -e
cd "$(dirname "$0")"
arq="$1"
nome="$(basename "${arq%.cl}")"
mkdir -p build
go run . --wat "$arq" > "build/$nome.wat"
wat2wasm "build/$nome.wat" -o "build/$nome.wasm"
exec wasmtime run "build/$nome.wasm"
