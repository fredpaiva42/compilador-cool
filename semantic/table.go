package semantic

import (
	"cool/ast"
	"fmt"
)

type ClassInfo struct {
	Name   string
	Parent string
	Node   *ast.Class
}

type Table struct {
	Classes map[string]*ClassInfo
	Methods map[string]map[string]*MethodSig
	Attrs   map[string]map[string]*AttrInfo
}

// Build monta a tabela a partir do programa.
// Nesta parte ela garante duas coisas:
//  1. nenhuma classe declarada duas vezes;
//  2. todo "inherits X" aponta para uma classe que existe
//     (exceto o vazio, que é o Object implícito e sempre existe).
func Build(prog *ast.Program) (*Table, error) {
	t := &Table{Classes: map[string]*ClassInfo{}}
	t.Classes["Object"] = &ClassInfo{Name: "Object", Parent: ""}

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

// checkCycles verifica se a herança forma árvore (sem ciclo, §3.2).
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
