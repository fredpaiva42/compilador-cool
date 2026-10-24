package semantic

import (
	"cool/ast"
	"fmt"
)

// ClassInfo é a entrada de uma classe na tabela; Node é nil nas básicas.
type ClassInfo struct {
	Name   string
	Parent string // "" quando não há inherits: o pai é Object
	Node   *ast.Class
}

// Table guarda a hierarquia e, para cada classe, métodos e atributos com os
// herdados já copiados.
type Table struct {
	Classes map[string]*ClassInfo
	Methods map[string]map[string]*MethodSig // classe → método → assinatura
	Attrs   map[string]map[string]*AttrInfo  // classe → atributo → ficha
	Types   map[ast.Expr]string
}

// Build monta a tabela de classes e valida o esqueleto da hierarquia:
// nenhuma classe declarada duas vezes, todo "inherits X" aponta para uma
// classe que existe (sem inherits, o pai é Object) e não há ciclo.
func Build(prog *ast.Program) (*Table, error) {
	t := &Table{Classes: map[string]*ClassInfo{}}
	t.Classes["Object"] = &ClassInfo{Name: "Object", Parent: ""}

	// Básicas da linguagem: entram na tabela sem nó, porque não são
	// declaradas no arquivo do programa.
	for _, b := range []string{"IO", "Int", "String", "Bool"} {
		t.Classes[b] = &ClassInfo{Name: b, Parent: ""}
	}

	for _, c := range prog.Classes {
		if _, dup := t.Classes[c.Name]; dup {
			return nil, fmt.Errorf("classe %s redefinida (linha %d)", c.Name, c.Line)
		}
		t.Classes[c.Name] = &ClassInfo{Name: c.Name, Parent: c.Parent, Node: c}
	}

	for _, c := range prog.Classes {
		if c.Parent == "" {
			continue // herda Object, que sempre existe
		}

		if _, ok := t.Classes[c.Parent]; !ok {
			return nil, fmt.Errorf("classes %s herda de %s inexistente (linha %d)", c.Name, c.Parent, c.Line)
		}
	}

	if err := t.checkCycles(); err != nil {
		return nil, err
	}

	return t, nil
}

// checkCycles verifica se a herança forma árvore (sem ciclo).
// Para cada classe, caminha para cima anotando o caminho;
// repetir um nome no mesmo caminho prova o ciclo.
func (t *Table) checkCycles() error {
	for name := range t.Classes {
		seen := map[string]bool{}
		cur := name
		for cur != "" {
			if seen[cur] {
				return fmt.Errorf("ciclo na herança envolvendo a classe %s", cur)
			}
			seen[cur] = true
			cur = t.Classes[cur].Parent
		}
	}
	return nil
}
