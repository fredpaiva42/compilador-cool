package parser

import (
	"testing"

	"cool/ast"
)

func TestIf(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { if x then 1 else 2 fi }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	n, ok := m.Body.(*ast.If)
	if !ok {
		t.Fatalf("deveria ser If: %+v", m.Body)
	}
	if o, ok := n.Cond.(*ast.Object); !ok || o.Name != "x" {
		t.Fatalf("cond errada: %+v", n.Cond)
	}
	if v, ok := n.Then.(*ast.IntConst); !ok || v.Value != "1" {
		t.Fatalf("then errado: %+v", n.Then)
	}
	if v, ok := n.Else.(*ast.IntConst); !ok || v.Value != "2" {
		t.Fatalf("else errado: %+v", n.Else)
	}
}

func TestWhile(t *testing.T) {
	prog := mustParse(t, `class P { f() : Object { while x loop y pool }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	if _, ok := m.Body.(*ast.While); !ok {
		t.Fatalf("deveria ser While: %+v", m.Body)
	}
}

func TestBlock(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { { 1; 2; } }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	b, ok := m.Body.(*ast.Block)
	if !ok || len(b.Exprs) != 2 {
		t.Fatalf("bloco com 2 exprs errado: %+v", m.Body)
	}
}

func TestBlockEmptyFails(t *testing.T) {
	if _, err := New(tokenize(`class P { f() : Int { { } }; };`)).ParseProgram(); err == nil {
		t.Fatal("bloco vazio deveria falhar")
	}
}

func TestLet(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { let x : Int <- 1, y : String in x }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	l, ok := m.Body.(*ast.Let)
	if !ok || len(l.Bindings) != 2 {
		t.Fatalf("let com 2 bindings errado: %+v", m.Body)
	}
	if !l.Bindings[0].HasInit || l.Bindings[1].HasInit {
		t.Fatal("flags HasInit errados")
	}
}

func TestNestedControl(t *testing.T) {
	prog := mustParse(t, `class P { f() : P { if isvoid x then new P else x fi }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	n, ok := m.Body.(*ast.If)
	if !ok {
		t.Fatalf("deveria ser If: %+v", m.Body)
	}
	if _, ok := n.Cond.(*ast.IsVoid); !ok {
		t.Fatalf("cond deveria ser IsVoid: %+v", n.Cond)
	}
}
