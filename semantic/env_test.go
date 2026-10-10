package semantic

import (
	"cool/ast"
	"testing"
)

func animalDog() *ast.Program {
	return &ast.Program{Classes: []*ast.Class{
		{Name: "Animal", Line: 1, Features: []ast.Feature{
			&ast.Attribute{Name: "nome", Type: "String", Line: 2},
			&ast.Method{Name: "falar", ReturnType: "String", Line: 3},
		}},
		{Name: "Cachorro", Parent: "Animal", Line: 4, Features: []ast.Feature{
			&ast.Attribute{Name: "patas", Type: "Int", Line: 5},
			&ast.Method{Name: "falar", ReturnType: "String", Line: 6},
		}},
	}}
}

func mustBuild(t *testing.T, p *ast.Program) *Table {
	t.Helper()
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}

	if err := tab.BuildEnvs(); err != nil {
		t.Fatalf("BuildEnvs falhou: %v", err)
	}

	return tab
}

func TestInheritCopies(t *testing.T) {
	tab := mustBuild(t, animalDog())
	attrs := tab.Attrs["Cachorro"]
	if attrs["nome"] == nil || attrs["nome"].Owner != "Animal" {
		t.Fatalf("nome herdado errado: %+v", attrs["nome"])
	}
	if attrs["patas"] == nil || attrs["patas"].Owner != "Cachorro" {
		t.Fatalf("patas próprio errado: %+v", attrs["patas"])
	}
	m := tab.Methods["Cachorro"]["falar"]
	if m == nil || m.Owner != "Cachorro" {
		t.Fatalf("falar deveria ser o override do filho: %+v", m)
	}
}
func TestAttrRedefineFails(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "A", Line: 1, Features: []ast.Feature{
			&ast.Attribute{Name: "x", Type: "Int", Line: 2},
		}},
		{Name: "B", Parent: "A", Line: 3, Features: []ast.Feature{
			&ast.Attribute{Name: "x", Type: "Int", Line: 4},
		}},
	}}
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}
	if err := tab.BuildEnvs(); err == nil {
		t.Fatal("atributo herdado redefinido deveria falhar")
	}
}
func TestBuiltinRedefineFails(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "String", Line: 1},
	}}
	if _, err := Build(p); err == nil {
		t.Fatal("redefinir String deveria falhar já no Build")
	}
}

func TestInheritIntFails(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "A", Parent: "Int", Line: 1},
	}}
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}
	if err := tab.BuildEnvs(); err == nil {
		t.Fatal("herdar de Int deveria falhar")
	}
}
func TestBuiltinMethodsExist(t *testing.T) {
	tab := mustBuild(t, animalDog())
	if tab.Methods["String"]["concat"] == nil {
		t.Fatal("String.concat deveria existir")
	}
	if tab.Methods["IO"]["out_string"] == nil {
		t.Fatal("IO.out_string deveria existir")
	}
	if tab.Methods["Object"]["copy"] == nil {
		t.Fatal("Object.copy deveria existir")
	}
	if tab.Methods["IO"]["abort"] == nil {
		t.Fatal("IO deveria herdar abort de Object")
	}
}
