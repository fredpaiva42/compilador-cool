package semantic

import (
	"testing"

	"cool/ast"
)

// envForTest monta um ambiente na classe P com a variável x:Int.
func envForTest(t *testing.T) *typeEnv {
	t.Helper()
	p := &ast.Program{Classes: []*ast.Class{{Name: "P", Line: 1}}}
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}
	if err := tab.BuildEnvs(); err != nil {
		t.Fatalf("BuildEnvs falhou: %v", err)
	}
	return &typeEnv{tab: tab, current: "P", vars: map[string]string{"x": "Int"}}
}

func mustType(t *testing.T, e *typeEnv, x ast.Expr) string {
	t.Helper()
	typ, err := e.checkExpr(x)
	if err != nil {
		t.Fatalf("não deveria falhar: %v (%+v)", err, x)
	}
	return typ
}

func mustFail(t *testing.T, e *typeEnv, x ast.Expr) {
	t.Helper()
	if _, err := e.checkExpr(x); err == nil {
		t.Fatalf("deveria falhar: %+v", x)
	}
}

func TestCheckAtoms(t *testing.T) {
	e := envForTest(t)
	if got := mustType(t, e, &ast.IntConst{Value: "42", Line: 1}); got != "Int" {
		t.Fatalf("Int deu %s", got)
	}
	if got := mustType(t, e, &ast.Object{Name: "x", Line: 1}); got != "Int" {
		t.Fatalf("x deu %s", got)
	}
	if got := mustType(t, e, &ast.Object{Name: "self", Line: 1}); got != "SELF_TYPE" {
		t.Fatalf("self deu %s", got)
	}
	mustFail(t, e, &ast.Object{Name: "fantasma", Line: 1})
}

func TestCheckArith(t *testing.T) {
	e := envForTest(t)
	one := &ast.IntConst{Value: "1", Line: 1}
	two := &ast.IntConst{Value: "2", Line: 1}
	str := &ast.StringConst{Value: "oi", Line: 1}
	if got := mustType(t, e, &ast.BinOp{Op: "+", Left: one, Right: two, Line: 1}); got != "Int" {
		t.Fatalf("+ deu %s", got)
	}
	mustFail(t, e, &ast.BinOp{Op: "+", Left: one, Right: str, Line: 1})
	if got := mustType(t, e, &ast.BinOp{Op: "<", Left: one, Right: two, Line: 1}); got != "Bool" {
		t.Fatalf("< deu %s", got)
	}
	if got := mustType(t, e, &ast.BinOp{Op: "=", Left: one, Right: two, Line: 1}); got != "Bool" {
		t.Fatalf("= deu %s", got)
	}
	mustFail(t, e, &ast.BinOp{Op: "=", Left: one, Right: str, Line: 1})
}
