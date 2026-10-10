package semantic

import (
	"testing"

	"cool/lexer"
	"cool/parser"
	"cool/token"
)

// checkSrc roda o pipeline inteiro: lexer → parser → semântico.
func checkSrc(t *testing.T, src string) error {
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
		t.Fatalf("parse falhou: %v", err)
	}
	return CheckProgram(prog)
}

func TestGoodHello(t *testing.T) {
	if err := checkSrc(t, `class Main inherits IO { main() : Object { out_string("oi\n") }; };`); err != nil {
		t.Fatalf("programa bom não deveria falhar: %v", err)
	}
}

func TestBadArith(t *testing.T) {
	if err := checkSrc(t, `class Main { main() : Object { 1 + "oi" }; };`); err == nil {
		t.Fatal(`1 + "oi" deveria falhar`)
	}
}

func TestMissingMain(t *testing.T) {
	if err := checkSrc(t, `class P { };`); err == nil {
		t.Fatal("programa sem Main deveria falhar")
	}
}

func TestMainInheritedFails(t *testing.T) {
	src := `class A { main() : Object { self }; };
	        class Main inherits A { };`
	if err := checkSrc(t, src); err == nil {
		t.Fatal("main herdado deveria falhar")
	}
}
