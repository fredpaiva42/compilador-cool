package parser

import (
	"cool/ast"
	"strings"
	"testing"
)

func TestAttributesOnly(t *testing.T) {
	prog, err := New(tokenize("class P {x : Int; y : String; };")).ParseProgram()
	if err != nil {
		t.Fatalf("não deveria falhar: %v", err)
	}

	feats := prog.Classes[0].Features
	if len(feats) != 2 {
		t.Fatalf("esperava 2 features, deu %d", len(feats))
	}

	a, ok := feats[0].(*ast.Attribute)
	if !ok || a.Name != "x" || a.Type != "Int" || a.HasInit {
		t.Fatalf("atributo 0 errado: %+v", feats[0])
	}
}

func TestFeatureBadFollow(t *testing.T) {
	_, err := New(tokenize("class P { x <- 1; };")).ParseProgram()
	if err == nil || !strings.Contains(err.Error(), "'(' ou ':'") {
		t.Fatalf("esperava erro de '(' ou ':', deu %v", err)
	}
}
