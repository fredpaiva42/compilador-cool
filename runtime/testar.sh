#!/usr/bin/env bash
# Compila o runtime e confere a bateria contra a saida esperada.
set -e
cd "$(dirname "$0")"
wat2wasm rt.wat -o rt.wasm
wasmtime run rt.wasm < bateria.in > bateria.saida
if diff -u bateria.esperado bateria.saida; then
  echo "BATERIA OK"
else
  echo "BATERIA FALHOU (veja o diff acima)"
  exit 1
fi
