package semantic

import (
	"testing"

	"cool/ast"
)

func checkDecl(t *testing.T, p *ast.Program) error {
	t.Helper()
	tab, err := Build(p)
	if err != nil {
		t.Fatalf("Build falhou: %v", err)
	}
	if err := tab.BuildEnvs(); err != nil {
		t.Fatalf("BuildEnvs falhou: %v", err)
	}
	return tab.checkDeclaredTypes()
}

func TestDeclaredTypesOK(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "A", Line: 1, Features: []ast.Feature{
			&ast.Attribute{Name: "x", Type: "SELF_TYPE", Line: 2},
			&ast.Method{Name: "f", ReturnType: "SELF_TYPE", Line: 3, Formals: []*ast.Formal{
				{Name: "a", Type: "Int", Line: 3},
			}},
		}},
	}}
	if err := checkDecl(t, p); err != nil {
		t.Fatalf("declarações válidas não deveriam falhar: %v", err)
	}
}

func TestFormalSelfTypeFails(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "A", Line: 1, Features: []ast.Feature{
			&ast.Method{Name: "f", ReturnType: "Int", Line: 2, Formals: []*ast.Formal{
				{Name: "a", Type: "SELF_TYPE", Line: 2},
			}},
		}},
	}}
	if err := checkDecl(t, p); err == nil {
		t.Fatal("formal SELF_TYPE deveria falhar")
	}
}

func TestUnknownTypeFails(t *testing.T) {
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "A", Line: 1, Features: []ast.Feature{
			&ast.Attribute{Name: "x", Type: "Fantasma", Line: 2},
		}},
	}}
	if err := checkDecl(t, p); err == nil {
		t.Fatal("tipo inexistente deveria falhar")
	}
}

func TestConformsTo(t *testing.T) {
	tab := threePets(t) // Animal + Cachorro + Gato
	casos := []struct {
		sub, sup, cur string
		espera        bool
	}{
		{"SELF_TYPE", "SELF_TYPE", "Cachorro", true},
		{"SELF_TYPE", "Animal", "Cachorro", true},
		{"SELF_TYPE", "Cachorro", "Cachorro", true},
		{"SELF_TYPE", "Int", "Cachorro", false},
		{"Cachorro", "SELF_TYPE", "Cachorro", false},
		{"Cachorro", "Animal", "Cachorro", true},
	}
	for _, c := range casos {
		if got := tab.conformsTo(c.sub, c.sup, c.cur); got != c.espera {
			t.Errorf("conformsTo(%s,%s,%s) = %v, queria %v",
				c.sub, c.sup, c.cur, got, c.espera)
		}
	}
}
