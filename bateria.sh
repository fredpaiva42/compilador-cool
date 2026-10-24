#!/usr/bin/env bash
# Bateria final do compilador: formato, vet, testes, runtime e todos os exemplos.
# Uso: ./bateria.sh
# Sai 0 com BATERIA OK, ou 1 com BATERIA FALHOU e a lista do que quebrou.
set -u
cd "$(dirname "$0")"

pass=0
fail=0
falhas=""

ok() { pass=$((pass + 1)); }
no() { fail=$((fail + 1)); falhas="$falhas\n- $1"; printf 'FALHOU: %s\n' "$1"; }

for ferramenta in go wat2wasm wasmtime python3; do
  command -v "$ferramenta" >/dev/null || { echo "falta $ferramenta no PATH"; exit 1; }
done

# 1. formato
if [ -n "$(gofmt -l .)" ]; then
  no "gofmt: $(gofmt -l . | tr '\n' ' ')"
else
  ok
fi

# 2. vet
if go vet ./... >/dev/null 2>&1; then ok; else no "go vet"; fi

# 3. testes
if go test ./... -count=1 >/dev/null 2>&1; then ok; else no "go test"; fi

# 4. runtime
if (cd runtime && ./testar.sh >/dev/null 2>&1); then ok; else no "runtime/testar.sh"; fi

# 5. executar exemplos e comparar a saida
mkdir -p build
RUN="minimo arith controle objetos novo vazio despacho estatico proprio ola conta igual caso tipo"
RUN_IN="eco dobro"
for nome in $RUN $RUN_IN; do
  entrada="/dev/null"
  [ -f "examples/$nome.entrada" ] && entrada="examples/$nome.entrada"
  if go run . --wat "examples/$nome.cl" > "build/$nome.wat" 2>"build/$nome.erro" \
    && wat2wasm "build/$nome.wat" -o "build/$nome.wasm" 2>>"build/$nome.erro" \
    && wasmtime run "build/$nome.wasm" <"$entrada" >"build/$nome.saida" 2>>"build/$nome.erro"; then
    if cmp -s "examples/$nome.esperado" "build/$nome.saida"; then ok; else no "$nome (saida difere)"; fi
  else
    no "$nome (falha de compilacao ou execucao)"
  fi
done

# 6. exemplos que passam (sai 0 com OK)
OK="demo:4 hello:1 helloWorld:1 peculiaridades:3 sem_ok_escopos:4 sem_ok_heranca:4 sem_ok_selftype:3"
for par in $OK; do
  nome="${par%%:*}"
  classes="${par##*:}"
  if go run . "examples/$nome.cl" >/dev/null 2>"build/$nome.erro"; then
    if grep -q "OK: $classes classe(s)" "build/$nome.erro"; then ok; else no "$nome (veredito inesperado)"; fi
  else
    no "$nome (deveria passar)"
  fi
done

# 7. exemplos que falham (mensagem esperada no stderr)
ERRO="
attrs.cl:programa sem classe Main
expr1.cl:programa sem classe Main
nested.cl:Main sem método main próprio
keywords.cl:esperado OBJECTID
numbers.cl:esperado CLASS, encontrado INT_CONST
operators.cl:esperado CLASS, encontrado OBJECTID
recovery.cl:esperava expressão, encontrei SEMI
sem_erro_atributo.cl:redefine herdado
sem_erro_dispatch.cl:inexistente em Animal
sem_erro_override.cl:com assinatura diferente
string_err.cl:newline cru dentro de string
strings.cl:esperado CLASS, encontrado STR_CONST
symbols.cl:esperado CLASS, encontrado LBRACE
type_err.cl:exige Int dos dois lados
unterminated.cl:não fechado
"
echo "$ERRO" | while IFS=: read -r arq msg; do
  [ -z "$arq" ] && continue
  if go run . "examples/$arq" >/dev/null 2>"build/erro.tmp"; then
    echo "FALHOU-ERRO: $arq (deveria falhar)"
  elif grep -qF "$msg" "build/erro.tmp"; then
    echo "ok-erro: $arq"
  else
    echo "FALHOU-ERRO: $arq (mensagem inesperada)"
  fi
done >build/erros.log 2>&1
if grep -q "FALHOU-ERRO" build/erros.log; then
  while IFS= read -r linha; do
    case "$linha" in FALHOU-ERRO*) no "${linha#FALHOU-ERRO: }";; esac
  done <build/erros.log
else
  nok=$(grep -c "ok-erro" build/erros.log)
  pass=$((pass + nok))
fi

echo "----------------"
echo "passaram: $pass, falharam: $fail"
if [ "$fail" -eq 0 ]; then
  echo "BATERIA OK"
else
  printf 'falhas:%b\n' "$falhas"
  echo "BATERIA FALHOU"
  exit 1
fi
