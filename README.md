# Compilador COOL — léxico, sintático, semântico e geração de código

Lexer, parser, analisador semântico e gerador de código manuais (sem
flex/yacc) da linguagem COOL, escritos em Go para a disciplina de
Compiladores. Especificação: Cool Reference Manual (Alex Aiken).
Alvo da geração: WASM em formato texto (WAT), via `wat2wasm` + `wasmtime`.

## Uso

    go run . examples/demo.cl          # três fases; tokens no stdout, veredito no stderr
    go run . --ast examples/demo.cl    # + árvore em S-expressions no stdout
    go run . examples/type_err.cl      # exemplo de erro semântico
    go run . --wat examples/minimo.cl  # módulo WAT no stdout
    ./compilar.sh examples/minimo.cl   # compila e executa (esperado: 42)
    ./compilar.sh examples/ola.cl      # exemplo com strings e IO
    printf '3\n4\n' | ./compilar.sh examples/dobro.cl  # exemplo com entrada (esperado: 7)
    ./bateria.sh                       # bateria final: formato, testes e todos os exemplos

Os exemplos `sem_ok_*.cl` (herança, SELF_TYPE, escopos) precisam passar e os
`sem_erro_*.cl` (override, atributo herdado, dispatch estático) precisam falhar
com a mensagem registrada no comentário do arquivo; `TestExemplosSemanticos`
confere os dois lados.

O `main` executa o pipeline nesta ordem: léxico (coleta tokens), barreira
léxica (`ILLEGAL` aborta antes do parser), sintático (`ParseProgram`),
semântico (`Analyze`, que devolve a tabela e o tipo de cada expressão),
geração (`Generate`, só com `--wat`) e por fim `OK: N classe(s)` no stderr.

## Arquitetura

- `token/`: vocabulário, tabela de keywords
  case-insensitive com a guarda de `true`/`false`, nomes legíveis.
- `lexer/`: scanner manual com lookahead de 1 (`readChar` como único caminho
  de avanço, `peekChar` para ver o próximo). Descarta branco e comentário em laço.
  Comentários `(* *)` com contador de profundidade (aninhados). Strings com
  escapes e 4 erros (`\n` cru, `\0`, EOF, limite 1024). Erros viram `ILLEGAL`
  com recuperação por ressincronização.
- `ast/`: nós da árvore (estruturais e expressões) e impressora
  S-expression (`DumpProgram`, flag `--ast`).
- `parser/`: descendente recursivo: programa, classe, feature (a decisão é
  pelo token após o nome: `(` indica método, `:` indica atributo), cascata
  de precedência (`<-` associa à direita, `not`, comparações que não
  associam, `+ -`, `* /`, `~` e `isvoid`, dispatch `.` e `@` com
  encadeamento, primárias). Parêntese agrupa sem criar nó. `new` é primária.
  Recuperação em modo pânico (sincroniza em `;`, `}` e próxima classe).
  Cada método recupera as próprias chaves.
- `semantic/`: tabela de classes (duplicada, pai inexistente, ciclo),
  ambientes com herdados copiados, básicas da Seção 8 (`abort`, `type_name`
  e `copy` presentes em toda classe), override com assinatura idêntica
  (Seção 6), `conformsTo` (relação de conformidade) e `Join` (ancestral comum
  para `if` e `case`), `SELF_TYPE` ancorado na classe corrente, checker por
  forma de expressão e o contrato de `Main.main` próprio sem formais
  (Seção 9).
- `runtime/`: runtime WASM escrito à mão (`rt.wat`): mapa de memória,
  alocador bump (`$alloc`), saída (`$putstr`, `$putint`), entrada com buffer
  (`$read_line`, `$read_int` com descarte do resto da linha), caixas
  (`$new_int`, `$new_bool`, `$new_str_empty`), objetos (`$obj_new`,
  `$copy_obj`, `$abort`), strings (`$str_concat`, `$str_substr`
  com aborto fora do alcance, `$str_eq` por conteúdo) e IO
  (`$io_out_string`, `$io_out_int`, `$io_in_string`, `$io_in_int`).
  O trecho entre as marcas `>>> RUNTIME` e `<<< RUNTIME` é o prelúdio que o
  emissor cola no início de todo módulo (`embed.go`, testado); o resto do
  arquivo é a bateria do runtime (`testar.sh`, `bateria.in/.esperado`).
- `codegen/`: emissor de WAT em estilo linear (uma instrução por linha).
  Todo valor é um endereço i32 (0 é void) com cabeçalho de tag e tamanho:
  Int guarda o i64 em +8, Bool o i32 em +8, String o tamanho e os bytes, e
  objeto de usuário um campo de 8 bytes por atributo. Construtor `$C_new`
  por classe (padrões, depois inits com `self` no objeto em construção).
  Despacho dinâmico ramifica pela tag até a implementação do dono (trava
  `void` com `unreachable`); o estático chama direto. `case` testa os ramos
  na ordem pela lista de tags e o primeiro que casa vence; `type_name`
  resolve o nome pela tag. Literais vão empacotados em dados a partir de
  2048. Tipos das classes: Object=1, IO=2, Int=3, String=4, Bool=5, resto em
  ordem alfabética de 6 em diante.
- `examples/`: os de etapa (`demo`, `hello`, `peculiaridades`, erros
  léxicos/sintáticos, `sem_ok_*`, `sem_erro_*`) e os do codegen, que a
  bateria executa de verdade: aritmética (`minimo`, `arith`), controle
  (`controle`), objetos (`objetos`, `novo`, `vazio`), despacho
  (`despacho`, `estatico`, `proprio`), strings e IO (`ola`, `conta`,
  `eco`, `dobro`, `igual`) e `case`/`type_name` (`caso`, `tipo`).
  Cada executável tem `.esperado` (stdout byte a byte) e, quando lê entrada,
  `.entrada`.
- `compilar.sh`: `--wat` + `wat2wasm` + `wasmtime run` para um exemplo.
- `bateria.sh`: a bateria final — `gofmt`, `vet`, `go test`, bateria do
  runtime, os 16 executáveis (compara stdout, usa `.entrada` no stdin),
  os 7 que passam e os 15 que falham (substring da mensagem).

## Limitações conhecidas

Divisão por zero aborta com trap (o manual deixa indefinido); o menor i64
não pode ser negado (overflow); o heap cresce sem coleta; linha terminada
em CRLF mantém o `\r`. Despacho e `case` sobre `void`, `case` sem ramo e
`substr` fora do alcance abrem como erro de execução, como manda o manual.
