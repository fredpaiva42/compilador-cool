# Compilador COOL — etapas léxica, sintática e semântica

Lexer, parser e analisador semântico manuais (sem flex/yacc) da linguagem
COOL, escritos em Go para a disciplina de Compiladores.
Especificação: Cool Reference Manual (Alex Aiken), Seções 4 a 12 e Figura 1.

## Uso

    go run . examples/demo.cl          # três fases; tokens no stdout, veredito no stderr
    go run . --ast examples/demo.cl    # + árvore em S-expressions no stdout
    go run . examples/typ_err.cl       # exemplo de erro semântico
    go test ./...                      # suíte completa

O `main` executa o pipeline nesta ordem: léxico (coleta tokens), barreira
léxica (`ILLEGAL` aborta antes do parser), sintático (`ParseProgram`),
semântico (`CheckProgram`), e por fim `OK: N classe(s)` no stderr.

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
- `examples/`: `demo.cl` (programa completo, passa as três fases),
  `hello.cl` e `helloWorld.cl` (programas), `typ_err.cl` (erro semântico),
  `recovery.cl`, `string_err.cl` e `unterminated.cl` (erros léxicos e
  sintáticos), e fixtures de etapa (`nested`, `attrs`, `expr1`, `numbers`,
  `symbols`, `operators`, `keywords`, `strings`).

