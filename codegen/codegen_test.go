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

// parseAndCheck roda o pipeline do lexico ao semantico e devolve a AST conferida.
func parseAndCheck(t *testing.T, src string) *ast.Program {
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
	if err := semantic.CheckProgram(prog); err != nil {
		t.Fatalf("semantico: %v", err)
	}
	return prog
}

func TestMinimalModule(t *testing.T) {
	wat, err := Generate(parseAndCheck(t, "class Main { main() : Int { 42 }; };"))
	if err != nil {
		t.Fatalf("gerar: %v", err)
	}
	pieces := []string{
		"(module",
		"(func $Main_main (param $self i32) (result i64)",
		"(i64.const 42)",
		`(func (export "_start")`,
		"(call $putint (call $Main_main (i32.const 0)))",
		"(call $proc_exit (i32.const 0))",
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

func TestMainNonIntNotPrinted(t *testing.T) {
	wat, err := Generate(parseAndCheck(t, "class Main { main() : Bool { true }; };"))
	if err != nil {
		t.Fatalf("gerar: %v", err)
	}
	// o preludio contem a definicao de $putint, entao a checagem e no _start
	i := strings.Index(wat, `(func (export "_start")`)
	if i < 0 {
		t.Fatal("modulo sem _start")
	}
	if strings.Contains(wat[i:], "putint") {
		t.Error("resultado que nao e Int nao deve ser impresso")
	}
	if !strings.Contains(wat[i:], "(drop (call $Main_main") {
		t.Error("resultado descartado deveria aparecer como drop")
	}
}

func TestUnsupportedExpression(t *testing.T) {
	_, err := Generate(parseAndCheck(t, "class Main { main() : Int { 1 + 2 }; };"))
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
