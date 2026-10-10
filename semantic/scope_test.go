package semantic

import (
	"testing"

	"cool/ast"
)

func TestAssignBlock(t *testing.T) {
	e := envForTest(t) // P, vars {x:Int}
	ok := &ast.Assign{Target: "x", Value: &ast.IntConst{Value: "2", Line: 1}, Line: 1}
	if got := mustType(t, e, ok); got != "Int" {
		t.Fatalf("x <- 2 deu %s", got)
	}
	mustFail(t, e, &ast.Assign{Target: "x", Value: &ast.StringConst{Value: "oi", Line: 1}, Line: 1})
	mustFail(t, e, &ast.Assign{Target: "y", Value: &ast.IntConst{Value: "1", Line: 1}, Line: 1})
	mustFail(t, e, &ast.Assign{Target: "self", Value: &ast.IntConst{Value: "1", Line: 1}, Line: 1})

	b := &ast.Block{Exprs: []ast.Expr{
		&ast.IntConst{Value: "1", Line: 1},
		&ast.StringConst{Value: "oi", Line: 1},
	}, Line: 1}
	if got := mustType(t, e, b); got != "String" {
		t.Fatalf("bloco deu %s, queria String", got)
	}
}

func TestIfWhile(t *testing.T) {
	e := envForTest(t)
	tru := &ast.BoolConst{Value: true, Line: 1}
	one := &ast.IntConst{Value: "1", Line: 1}
	two := &ast.IntConst{Value: "2", Line: 1}
	n := &ast.If{Cond: tru, Then: one, Else: two, Line: 1}
	if got := mustType(t, e, n); got != "Int" {
		t.Fatalf("if deu %s", got)
	}
	mustFail(t, e, &ast.If{Cond: one, Then: one, Else: two, Line: 1})

	w := &ast.While{Cond: tru, Body: one, Line: 1}
	if got := mustType(t, e, w); got != "Object" {
		t.Fatalf("while deu %s, queria Object", got)
	}
	mustFail(t, e, &ast.While{Cond: one, Body: one, Line: 1})
}

func TestIfJoin(t *testing.T) {
	tab := threePets(t)
	e := &typeEnv{tab: tab, current: "Main", vars: map[string]string{}}
	n := &ast.If{
		Cond: &ast.BoolConst{Value: true, Line: 1},
		Then: &ast.New{Type: "Cachorro", Line: 1},
		Else: &ast.New{Type: "Gato", Line: 1},
		Line: 1,
	}
	if got := mustType(t, e, n); got != "Animal" {
		t.Fatalf("if Cachorro/Gato deu %s, queria Animal", got)
	}
}

func TestLet(t *testing.T) {
	e := envForTest(t)
	l := &ast.Let{
		Bindings: []*ast.LetBinding{
			{Name: "y", Type: "Int", Init: &ast.IntConst{Value: "1", Line: 1}, HasInit: true, Line: 1},
		},
		Body: &ast.Object{Name: "y", Line: 1},
		Line: 1,
	}
	if got := mustType(t, e, l); got != "Int" {
		t.Fatalf("let deu %s", got)
	}
	// init usa binding anterior.
	l2 := &ast.Let{
		Bindings: []*ast.LetBinding{
			{Name: "a", Type: "Int", Init: &ast.IntConst{Value: "1", Line: 1}, HasInit: true, Line: 1},
			{Name: "b", Type: "Int", Init: &ast.Object{Name: "a", Line: 1}, HasInit: true, Line: 1},
		},
		Body: &ast.Object{Name: "b", Line: 1},
		Line: 1,
	}
	if got := mustType(t, e, l2); got != "Int" {
		t.Fatalf("let encadeado deu %s", got)
	}
	// init com tipo errado.
	mustFail(t, e, &ast.Let{
		Bindings: []*ast.LetBinding{
			{Name: "z", Type: "String", Init: &ast.IntConst{Value: "1", Line: 1}, HasInit: true, Line: 1},
		},
		Body: &ast.Object{Name: "z", Line: 1},
		Line: 1,
	})
}
