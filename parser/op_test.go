package parser

import (
	"testing"

	"cool/ast"
)

func methodBody(t *testing.T, src string) ast.Expr {
	t.Helper()
	prog := mustParse(t, src)
	return prog.Classes[0].Features[0].(*ast.Method).Body
}

func TestArithPrecedence(t *testing.T) {
	b, ok := methodBody(t, `class P { f() : Int { 1 + 2 * 3 }; };`).(*ast.BinOp)
	if !ok || b.Op != "+" {
		t.Fatalf("topo deveria ser +: %+v", b)
	}
	r, ok := b.Right.(*ast.BinOp)
	if !ok || r.Op != "*" {
		t.Fatalf("direita deveria ser *: %+v", b.Right)
	}
}

func TestAssignRightAssoc(t *testing.T) {
	b, ok := methodBody(t, `class P { f() : Int { x <- y <- 2 }; };`).(*ast.Assign)
	if !ok || b.Target != "x" {
		t.Fatalf("externo deveria ser x <-: %+v", b)
	}
	inner, ok := b.Value.(*ast.Assign)
	if !ok || inner.Target != "y" {
		t.Fatalf("interno deveria ser y <-: %+v", b.Value)
	}
}

func TestAssignTargetMustBeIdent(t *testing.T) {
	if _, err := New(tokenize(`class P { f() : Int { 1 <- 2 }; };`)).ParseProgram(); err == nil {
		t.Fatal("alvo não-identificador deveria falhar")
	}
}

func TestComparisonNonAssoc(t *testing.T) {
	if _, err := New(tokenize(`class P { f() : Bool { a < b < c }; };`)).ParseProgram(); err == nil {
		t.Fatal("comparação encadeada deveria falhar")
	}
}

func TestNotBindsLooserThanComparison(t *testing.T) {
	n, ok := methodBody(t, `class P { f() : Bool { not x = y }; };`).(*ast.Not)
	if !ok {
		t.Fatalf("topo deveria ser Not: %+v", n)
	}
	if _, ok := n.Expr.(*ast.BinOp); !ok {
		t.Fatalf("operando deveria ser comparação: %+v", n.Expr)
	}
}

func TestNegBindsTighterThanAdd(t *testing.T) {
	b, ok := methodBody(t, `class P { f() : Int { ~x + 1 }; };`).(*ast.BinOp)
	if !ok || b.Op != "+" {
		t.Fatalf("topo deveria ser +: %+v", b)
	}
	if _, ok := b.Left.(*ast.Neg); !ok {
		t.Fatalf("esquerda deveria ser Neg: %+v", b.Left)
	}
}

func TestIsvoidTighterThanComparison(t *testing.T) {
	b, ok := methodBody(t, `class P { f() : Bool { isvoid x = y }; };`).(*ast.BinOp)
	if !ok {
		t.Fatalf("topo deveria ser comparação: %+v", b)
	}
	if _, ok := b.Left.(*ast.IsVoid); !ok {
		t.Fatalf("esquerda deveria ser IsVoid: %+v", b.Left)
	}
}
