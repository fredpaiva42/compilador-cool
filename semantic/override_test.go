package semantic

import (
	"cool/ast"
	"testing"
)

func TestOverrideOK(t *testing.T) {
	mustBuild(t, animalDog()) // falar():String nos dois: passa
}

func TestOverrideBadReturn(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "P", Line: 1, Features: []ast.Feature{
			&ast.Method{Name: "f", ReturnType: "Int", Line: 2},
		}},
		{Name: "C", Parent: "P", Line: 3, Features: []ast.Feature{
			&ast.Method{Name: "f", ReturnType: "String", Line: 4},
		}},
	}}

	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}

	if err := tab.BuildEnvs(); err == nil {
		t.Fatalf("retorno diferente deveria falhar")
	}
}

func TestOverrideBadArity(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "P", Line: 1, Features: []ast.Feature{
			&ast.Method{Name: "f", ReturnType: "Int", Line: 2, Formals: []*ast.Formal{
				{Name: "x", Type: "Int"},
			}},
		}},
		{Name: "C", Parent: "P", Line: 3, Features: []ast.Feature{
			&ast.Method{Name: "f", ReturnType: "Int", Line: 4},
		}},
	}}

	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}

	if err := tab.BuildEnvs(); err == nil {
		t.Fatalf("aridade diferente deveria falhar")
	}
}
