package parser

import (
	"testing"

	"cool/ast"
)

func TestImplicitDispatch(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { g() }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	d, ok := m.Body.(*ast.Dispatch)
	if !ok || d.Receiver != nil || d.Method != "g" || len(d.Args) != 0 {
		t.Fatalf("dispatch implícito errado: %+v", m.Body)
	}
}

func TestDynamicDispatchArgs(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { x.concat(y) }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	d, ok := m.Body.(*ast.Dispatch)
	if !ok || d.Method != "concat" || len(d.Args) != 1 {
		t.Fatalf("dispatch dinâmico errado: %+v", m.Body)
	}
	recv, ok := d.Receiver.(*ast.Object)
	if !ok || recv.Name != "x" {
		t.Fatalf("receptor errado: %+v", d.Receiver)
	}
}

func TestChainedDispatch(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { a.g().h() }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	outer, ok := m.Body.(*ast.Dispatch)
	if !ok || outer.Method != "h" {
		t.Fatalf("externo errado: %+v", m.Body)
	}
	if _, ok := outer.Receiver.(*ast.Dispatch); !ok {
		t.Fatalf("interno deveria ser Dispatch: %+v", outer.Receiver)
	}
}

func TestStaticDispatch(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int { x@A.g(1) }; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	s, ok := m.Body.(*ast.StaticDispatch)
	if !ok || s.StaticType != "A" || s.Method != "g" || len(s.Args) != 1 {
		t.Fatalf("dispatch estático errado: %+v", m.Body)
	}
}
