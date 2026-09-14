package parser

import (
	"cool/ast"
	"strings"
	"testing"
)

func TestPanicModeTwoErrors(t *testing.T) {
	src := `class A {
	  x : Int <- 1 +;
	  y : Bool <- true;
	};
	class B {
	  f() : Int { 2 * };
	};`
	prog, err := New(tokenize(src)).ParseProgram()

	if err == nil {
		t.Fatal("esperava erros, deu nil")
	}
	if got := strings.Count(err.Error(), "\n") + 1; got != 2 {
		t.Fatalf("esperava exatamente 2 erros, deu %d:\n%s", got, err)
	}

	if !strings.Contains(err.Error(), "esperava expressão") {
		t.Fatalf("faltou o erro do operando ausente:\n%s", err)
	}

	if len(prog.Classes) != 2 {
		t.Fatalf("esperava 2 classes parciais, deu %d", len(prog.Classes))
	}

	if len(prog.Classes[0].Features) != 1 {
		t.Fatalf("feature boa de A deveria sobreviver, deu %d", len(prog.Classes[0].Features))
	}

	m, ok := prog.Classes[1].Features[0].(*ast.Method)
	if !ok || m.Body != nil {
		t.Fatalf("metodo de B deveria sobreviver parcial (corpo nil): %+v", prog.Classes[1].Features[0])
	}

}
