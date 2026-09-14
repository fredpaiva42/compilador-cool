package parser

import (
	"testing"

	"cool/ast"
)

func TestCaseBasic(t *testing.T) {
	prog := mustParse(t, `class P { f() : Int {
	  case x of
	    a : A => 1;
	    b : B => 2;
	  esac
	}; };`)
	m := prog.Classes[0].Features[0].(*ast.Method)
	c, ok := m.Body.(*ast.Case)
	if !ok || len(c.Branches) != 2 {
		t.Fatalf("case com 2 ramos errado: %+v", m.Body)
	}
	if s, ok := c.Subject.(*ast.Object); !ok || s.Name != "x" {
		t.Fatalf("alvo errado: %+v", c.Subject)
	}
	if c.Branches[0].Name != "a" || c.Branches[0].Type != "A" {
		t.Fatalf("ramo 0 errado: %+v", c.Branches[0])
	}
	if v, ok := c.Branches[1].Body.(*ast.IntConst); !ok || v.Value != "2" {
		t.Fatalf("corpo do ramo 1 errado: %+v", c.Branches[1].Body)
	}
}

func TestCaseEmptyFails(t *testing.T) {
	if _, err := New(tokenize(`class P { f() : Int { case x of esac }; };`)).ParseProgram(); err == nil {
		t.Fatal("case sem ramos deveria falhar")
	}
}

func TestCaseMissingArrowFails(t *testing.T) {
	if _, err := New(tokenize(`class P { f() : Int { case x of a : A 1; esac }; };`)).ParseProgram(); err == nil {
		t.Fatal("ramo sem '=>' deveria falhar")
	}
}
