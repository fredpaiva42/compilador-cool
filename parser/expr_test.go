package parser

import (
	"testing"

	"cool/ast"
)

func mustParse(t *testing.T, src string) *ast.Program {
	t.Helper()
	prog, err := New(tokenize(src)).ParseProgram()
	if err != nil {
		t.Fatalf("não deveria falhar: %v", err)
	}
	return prog
}

func TestPrimaries(t *testing.T) {
	prog := mustParse(t, `class P {
	  a : Int <- 42;
	  b : String <- "oi";
	  c : Bool <- tRue;
	  d : Int <- x;
	};`)
	feats := prog.Classes[0].Features
	if len(feats) != 4 {
		t.Fatalf("esperava 4 atributos, deu %d", len(feats))
	}
	a := feats[0].(*ast.Attribute)
	if n, ok := a.Init.(*ast.IntConst); !ok || n.Value != "42" {
		t.Fatalf("init 0 errado: %+v", a.Init)
	}
	b := feats[1].(*ast.Attribute)
	if s, ok := b.Init.(*ast.StringConst); !ok || s.Value != "oi" {
		t.Fatalf("init 1 errado: %+v", b.Init)
	}
	c := feats[2].(*ast.Attribute)
	if bl, ok := c.Init.(*ast.BoolConst); !ok || !bl.Value {
		t.Fatalf("init 2 errado (tRue deveria ser true): %+v", c.Init)
	}
	d := feats[3].(*ast.Attribute)
	if o, ok := d.Init.(*ast.Object); !ok || o.Name != "x" {
		t.Fatalf("init 3 errado: %+v", d.Init)
	}
}

func TestNewIsVoid(t *testing.T) {
	prog := mustParse(t, `class P {
	  f() : Bool { isvoid x };
	  g() : P { new P };
	  h() : Bool { isvoid isvoid x };
	};`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	if _, ok := m.Body.(*ast.IsVoid); !ok {
		t.Fatalf("corpo de f deveria ser IsVoid: %+v", m.Body)
	}
	g := prog.Classes[0].Features[1].(*ast.Method)
	if n, ok := g.Body.(*ast.New); !ok || n.Type != "P" {
		t.Fatalf("corpo de g deveria ser New P: %+v", g.Body)
	}
	h := prog.Classes[0].Features[2].(*ast.Method)
	outer, ok := h.Body.(*ast.IsVoid)
	if !ok {
		t.Fatalf("corpo de h deveria ser IsVoid: %+v", h.Body)
	}
	if _, ok := outer.Expr.(*ast.IsVoid); !ok {
		t.Fatalf("operando de h deveria ser IsVoid aninhado: %+v", outer.Expr)
	}
}

func TestParenGroups(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { (42) }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	if n, ok := m.Body.(*ast.IntConst); !ok || n.Value != "42" {
		t.Fatalf("parêntese deveria devolver o de dentro: %+v", m.Body)
	}
}
