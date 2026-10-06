package semantic

import (
	"cool/ast"
	"testing"
)

func prog(classes ...*ast.Class) *ast.Program {
	return &ast.Program{Classes: classes}
}

func TestTableOk(t *testing.T) {
	p := prog(
		&ast.Class{Name: "Animal", Parent: "", Line: 1},
		&ast.Class{Name: "Dog", Parent: "Animal", Line: 2},
	)
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("não deveria falhar: %v", err)
	}

	if _, ok := tab.Classes["Object"]; !ok {
		t.Fatal("Object deveria existir na tabela")
	}

	if tab.Classes["Dog"].Parent != "Animal" {
		t.Fatal("pai de Dog errado")
	}
}

func TestTableDup(t *testing.T) {
	p := prog(
		&ast.Class{Name: "A", Parent: "", Line: 1},
		&ast.Class{Name: "A", Parent: "", Line: 2},
	)
	if _, err := Build(p); err == nil {
		t.Fatal("classe duplicada deveria falhar")
	}
}
func TestTableParentMissing(t *testing.T) {
	p := prog(&ast.Class{Name: "A", Parent: "Fantasma", Line: 1})
	if _, err := Build(p); err == nil {
		t.Fatal("pai inexistente deveria falhar")
	}
}
func TestTableSelfCycle(t *testing.T) {
	p := prog(&ast.Class{Name: "A", Parent: "A", Line: 1})
	if _, err := Build(p); err == nil {
		t.Fatal("A herda de A deveria falhar")
	}
}
func TestTableTwoCycle(t *testing.T) {
	p := prog(
		&ast.Class{Name: "A", Parent: "B", Line: 1},
		&ast.Class{Name: "B", Parent: "A", Line: 2},
	)
	if _, err := Build(p); err == nil {
		t.Fatal("ciclo A<->B deveria falhar")
	}
}
