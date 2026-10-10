package semantic

import (
	"testing"

	"cool/ast"
)

// threePets: Animal + dois filhos (para o join ter onde se encontrar).
func threePets(t *testing.T) *Table {
	t.Helper()
	p := &ast.Program{Classes: []*ast.Class{
		{Name: "Animal", Line: 1},
		{Name: "Cachorro", Parent: "Animal", Line: 2},
		{Name: "Gato", Parent: "Animal", Line: 3},
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

func TestConforms(t *testing.T) {
	tab := threePets(t)
	casos := []struct {
		sub, sup string
		espera   bool
	}{
		{"Cachorro", "Animal", true},
		{"Cachorro", "Object", true},
		{"Animal", "Animal", true},
		{"Animal", "Cachorro", false},
		{"Int", "Object", true},
		{"String", "Int", false},
	}
	for _, c := range casos {
		if got := tab.Conforms(c.sub, c.sup); got != c.espera {
			t.Errorf("Conforms(%s,%s) = %v, queria %v", c.sub, c.sup, got, c.espera)
		}
	}
}

func TestJoin(t *testing.T) {
	tab := threePets(t)
	if got := tab.Join("Cachorro", "Gato"); got != "Animal" {
		t.Errorf("Join(Cachorro,Gato) = %s, queria Animal", got)
	}
	if got := tab.Join("Cachorro", "Animal"); got != "Animal" {
		t.Errorf("Join(Cachorro,Animal) = %s, queria Animal", got)
	}
	if got := tab.Join("Int", "String"); got != "Object" {
		t.Errorf("Join(Int,String) = %s, queria Object", got)
	}
	if got := tab.Join("Gato", "Gato"); got != "Gato" {
		t.Errorf("Join(Gato,Gato) = %s, queria Gato", got)
	}
	if got := tab.JoinAll([]string{"Cachorro", "Gato", "Animal"}); got != "Animal" {
		t.Errorf("JoinAll = %s, queria Animal", got)
	}
}
