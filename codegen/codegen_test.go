package codegen

import (
	"strings"
	"testing"

	"cool/ast"
	"cool/lexer"
	"cool/parser"
	"cool/semantic"
	"cool/token"
)

// parseAndCheck roda o pipeline do lexico ao semantico e devolve a AST e a analise.
func parseAndCheck(t *testing.T, src string) (*ast.Program, *semantic.Result) {
	t.Helper()
	l := lexer.New(src)
	var toks []token.Token
	for {
		tok := l.NextToken()
		toks = append(toks, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	prog, err := parser.New(toks).ParseProgram()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := semantic.Analyze(prog)
	if err != nil {
		t.Fatalf("semantico: %v", err)
	}
	return prog, res
}

// generate roda o emissor inteiro sobre um programa COOL.
func generate(t *testing.T, src string) string {
	t.Helper()
	prog, res := parseAndCheck(t, src)
	wat, err := Generate(prog, res)
	if err != nil {
		t.Fatalf("gerar: %v", err)
	}
	return wat
}

// ops devolve so as instrucoes do corpo de Main.main, sem a declaracao da
// funcao nem o fechamento. O preludio do runtime nao entra: ele e fixo.
func ops(t *testing.T, src string) []string {
	t.Helper()
	wat := generate(t, src)
	i := strings.Index(wat, "(func $Main_main")
	if i < 0 {
		t.Fatalf("modulo sem Main.main:\n%s", wat)
	}
	j := strings.Index(wat[i:], "\n  )")
	if j < 0 {
		t.Fatalf("nao achei o fim de Main.main:\n%s", wat)
	}
	var out []string
	for _, ln := range strings.Split(wat[i:i+j], "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "(func") {
			continue
		}
		out = append(out, ln)
	}
	return out
}

// checkOps compara a sequencia de instrucoes geradas com a esperada.
func checkOps(t *testing.T, src string, want ...string) {
	t.Helper()
	got := ops(t, src)
	if len(got) != len(want) {
		t.Fatalf("instrucoes = %v, quer %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("instrucoes = %v, quer %v", got, want)
		}
	}
}

func TestMinimalModule(t *testing.T) {
	wat := generate(t, "class Main { main() : Int { 42 }; };")
	pieces := []string{
		"(module",
		"(func $Main_main (param $self i32) (result i64)",
		"i64.const 42",
		`(func (export "_start")`,
		"call $Main_main",
		"call $putint",
		"call $proc_exit",
	}
	for _, piece := range pieces {
		if !strings.Contains(wat, piece) {
			t.Errorf("modulo sem %q\n--- modulo ---\n%s", piece, wat)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(wat), ")") {
		t.Error("modulo nao termina fechando o parentese")
	}
}

func TestArithmeticPrecedence(t *testing.T) {
	// 1 + 2 * 3 empilha o 1, depois o 2, o 3 e a multiplicacao, e so no fim soma
	checkOps(t, "class Main { main() : Int { 1 + 2 * 3 }; };",
		"i64.const 1", "i64.const 2", "i64.const 3", "i64.mul", "i64.add")
}

func TestArithmeticDivisionAndSubtraction(t *testing.T) {
	checkOps(t, "class Main { main() : Int { 100 / 2 - 8 }; };",
		"i64.const 100", "i64.const 2", "i64.div_s", "i64.const 8", "i64.sub")
}

func TestComparison(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { 1 < 2 }; };",
		"i64.const 1", "i64.const 2", "i64.lt_s")
	checkOps(t, "class Main { main() : Bool { 1 <= 2 }; };",
		"i64.const 1", "i64.const 2", "i64.le_s")
}

func TestNegIsSubtraction(t *testing.T) {
	// o WASM nao tem negacao de inteiro
	checkOps(t, "class Main { main() : Int { ~5 }; };",
		"i64.const 0", "i64.const 5", "i64.sub")
}

func TestNotIsEqz(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { not true }; };",
		"i32.const 1", "i32.eqz")
}

func TestEqualityInt(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { 1 = 2 }; };",
		"i64.const 1", "i64.const 2", "i64.eq")
}

func TestEqualityBool(t *testing.T) {
	// Bool e i32, entao a comparacao e i32.eq
	checkOps(t, "class Main { main() : Bool { true = false }; };",
		"i32.const 1", "i32.const 0", "i32.eq")
}

func TestBlockDropsAllButLast(t *testing.T) {
	checkOps(t, "class Main { main() : Int { { 1; 2; 3; } }; };",
		"i64.const 1", "drop", "i64.const 2", "drop", "i64.const 3")
}

func TestMainNonIntNotPrinted(t *testing.T) {
	wat := generate(t, "class Main { main() : Bool { true }; };")
	// o preludio contem a definicao de $putint, entao a checagem e no _start
	i := strings.Index(wat, `(func (export "_start")`)
	if i < 0 {
		t.Fatal("modulo sem _start")
	}
	if strings.Contains(wat[i:], "putint") {
		t.Error("resultado que nao e Int nao deve ser impresso")
	}
	if !strings.Contains(wat[i:], "drop") {
		t.Error("resultado descartado deveria aparecer como drop")
	}
}

func TestUnsupportedExpression(t *testing.T) {
	prog, res := parseAndCheck(t, `class Main { main() : Object { "oi" }; };`)
	_, err := Generate(prog, res)
	if err == nil {
		t.Fatal("esperava erro para expressao ainda nao suportada")
	}
	if !strings.Contains(err.Error(), "nao suportada") {
		t.Errorf("mensagem inesperada: %v", err)
	}
}

func TestWATType(t *testing.T) {
	cases := map[string]string{
		"Int": "i64", "Bool": "i32", "String": "i32",
		"Object": "i32", "IO": "i32", "SELF_TYPE": "i32",
	}
	for cool, want := range cases {
		if got := watType(cool); got != want {
			t.Errorf("watType(%q) = %q, quer %q", cool, got, want)
		}
	}
}

func TestIfThenElse(t *testing.T) {
	checkOps(t, "class Main { main() : Int { if 1 < 2 then 10 else 20 fi }; };",
		"i64.const 1", "i64.const 2", "i64.lt_s",
		"(if (result i64)",
		"(then", "i64.const 10", ")",
		"(else", "i64.const 20", ")",
		")")
}

func TestLetAssign(t *testing.T) {
	checkOps(t, "class Main { main() : Int { let x : Int <- 40 in { x <- x + 2; } }; };",
		"(local $l0 i64)",
		"i64.const 40", "local.set $l0",
		"local.get $l0", "i64.const 2", "i64.add", "local.tee $l0")
}

func TestLetDefault(t *testing.T) {
	checkOps(t, "class Main { main() : Int { let x : Int in x + 1 }; };",
		"(local $l0 i64)",
		"i64.const 0", "local.set $l0",
		"local.get $l0", "i64.const 1", "i64.add")
}

func TestLetShadowing(t *testing.T) {
	// o interno ganha outro local e o externo volta a valer depois
	checkOps(t, "class Main { main() : Int { let x : Int <- 1 in { { let x : Int <- 2 in x; }; x; } }; };",
		"(local $l0 i64)", "(local $l1 i64)",
		"i64.const 1", "local.set $l0",
		"i64.const 2", "local.set $l1",
		"local.get $l1", "drop",
		"local.get $l0")
}

func TestIsVoidIntIsFalse(t *testing.T) {
	// Int nunca e void: o valor sai da pilha e entra falso
	checkOps(t, "class Main { main() : Int { if isvoid 0 then 1 else 2 fi }; };",
		"i64.const 0", "drop", "i32.const 0",
		"(if (result i64)",
		"(then", "i64.const 1", ")",
		"(else", "i64.const 2", ")",
		")")
}

func TestSelfIsParam(t *testing.T) {
	checkOps(t, "class Main { main() : Object { self }; };",
		"local.get $self")
}

func TestWhileEmission(t *testing.T) {
	checkOps(t, "class Main { main() : Object { while true loop 1 pool }; };",
		"(block $w0", "(loop $l0",
		"i32.const 1", "i32.eqz", "br_if $w0",
		"i64.const 1", "drop", "br $l0",
		")", ")",
		"i32.const 0")
}

func TestIfMixedBranchesError(t *testing.T) {
	// join(Int, Bool) e Object: sem boxing o emissor recusa
	prog, res := parseAndCheck(t, "class Main { main() : Object { if true then 1 else false fi }; };")
	_, err := Generate(prog, res)
	if err == nil {
		t.Fatal("esperava erro para ramos de tipos diferentes")
	}
	if !strings.Contains(err.Error(), "ainda nao suportado") {
		t.Errorf("mensagem inesperada: %v", err)
	}
}

func TestAttributeNotYetSupported(t *testing.T) {
	prog, res := parseAndCheck(t, "class Main { x : Int <- 5; main() : Int { x }; };")
	_, err := Generate(prog, res)
	if err == nil {
		t.Fatal("esperava erro para atributo")
	}
	if !strings.Contains(err.Error(), "atributo") {
		t.Errorf("mensagem inesperada: %v", err)
	}
}
