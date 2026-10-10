package semantic

import (
	"testing"

	"cool/ast"
)

// dispatchTab: Animal{falar():String, comer(c:String):String} + Cachorro herda.
func dispatchTab(t *testing.T) *Table {
	t.Helper()
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "Animal", Line: 1, Features: []ast.Feature{
			&ast.Method{Name: "falar", ReturnType: "String", Line: 2},
			&ast.Method{Name: "comer", ReturnType: "String", Line: 3, Formals: []*ast.Formal{
				{Name: "c", Type: "String", Line: 3},
			}},
		}},
		{Name: "Cachorro", Parent: "Animal", Line: 4},
	}}
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}
	if err := tab.BuildEnvs(); err != nil {
		t.Fatalf("BuildEnvs falhou: %v", err)
	}
	return tab
}

func TestDynamicBasic(t *testing.T) {
	tab := dispatchTab(t)
	e := &typeEnv{tab: tab, current: "Main", vars: map[string]string{"b": "Animal"}}
	d := &ast.Dispatch{
		Receiver: &ast.Object{Name: "b", Line: 1},
		Method:   "falar", Line: 1,
	}
	if got := mustType(t, e, d); got != "String" {
		t.Fatalf("b.falar() deu %s", got)
	}
}

func TestDispatchArgs(t *testing.T) {
	tab := dispatchTab(t)
	e := &typeEnv{tab: tab, current: "Main", vars: map[string]string{}}
	ok := &ast.Dispatch{
		Receiver: &ast.StringConst{Value: "oi", Line: 1},
		Method:   "concat", Line: 1,
		Args: []ast.Expr{&ast.StringConst{Value: "!", Line: 1}},
	}
	if got := mustType(t, e, ok); got != "String" {
		t.Fatalf("concat deu %s", got)
	}
	bad := &ast.Dispatch{
		Receiver: &ast.StringConst{Value: "oi", Line: 1},
		Method:   "concat", Line: 1,
		Args: []ast.Expr{&ast.IntConst{Value: "1", Line: 1}},
	}
	mustFail(t, e, bad)
	mustFail(t, e, &ast.Dispatch{
		Receiver: &ast.Object{Name: "self", Line: 1},
		Method:   "inexistente", Line: 1,
	})
}

func TestSelfTypeReturn(t *testing.T) {
	tab := dispatchTab(t)
	// copy() de Object devolve SELF_TYPE: receptor Animal → Animal.
	e := &typeEnv{tab: tab, current: "Main", vars: map[string]string{"b": "Animal"}}
	d := &ast.Dispatch{
		Receiver: &ast.Object{Name: "b", Line: 1},
		Method:   "copy", Line: 1,
	}
	if got := mustType(t, e, d); got != "Animal" {
		t.Fatalf("b.copy() deu %s, queria Animal", got)
	}
	// receptor self → SELF_TYPE preservado.
	e2 := &typeEnv{tab: tab, current: "Animal", vars: map[string]string{}}
	d2 := &ast.Dispatch{Receiver: nil, Method: "copy", Line: 1}
	if got := mustType(t, e2, d2); got != "SELF_TYPE" {
		t.Fatalf("copy() implícito deu %s, queria SELF_TYPE", got)
	}
}

func TestStaticDispatch(t *testing.T) {
	tab := dispatchTab(t)
	e := &typeEnv{tab: tab, current: "Main", vars: map[string]string{"g": "Cachorro"}}
	ok := &ast.StaticDispatch{
		Receiver:   &ast.Object{Name: "g", Line: 1},
		StaticType: "Animal", Method: "falar", Line: 1,
	}
	if got := mustType(t, e, ok); got != "String" {
		t.Fatalf("g@Animal.falar() deu %s", got)
	}
	// receptor não conforma com o escrito.
	mustFail(t, e, &ast.StaticDispatch{
		Receiver:   &ast.Object{Name: "g", Line: 1},
		StaticType: "String", Method: "length", Line: 1,
	})
}
