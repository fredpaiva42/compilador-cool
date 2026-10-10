package semantic

import (
	"cool/ast"
	"fmt"
)

// checkDeclaredTypes garante que todo tipo declarado existe
// e que SELF_TYPE só aparece onde é permitido.
// Cobre atributos, retornos e formais.
func (t *Table) checkDeclaredTypes() error {
	known := func(typ string) bool {
		if typ == "SELF_TYPE" {
			return true
		}

		_, ok := t.Classes[typ]
		return ok
	}

	for _, info := range t.Classes {
		if info.Node == nil {
			continue
		}

		for _, f := range info.Node.Features {
			switch n := f.(type) {
			case *ast.Attribute:
				if !known(n.Type) {
					return fmt.Errorf("atributo %s de %s declara tipo inexistente %s (linha %d)", n.Name, info.Name, n.Type, n.Line)
				}
			case *ast.Method:
				if !known(n.ReturnType) {
					return fmt.Errorf("método %s de %s declara retorno inexistente %s (linha %d)", n.Name, info.Name, n.ReturnType, n.Line)
				}

				for _, fm := range n.Formals {
					if fm.Type == "SELF_TYPE" {
						return fmt.Errorf("formal %s de %s.%s não pode ser SELF_TYPE (linha %d)", fm.Name, info.Name, n.Name, fm.Line)
					}

					if !known(fm.Type) {
						return fmt.Errorf("formal %s de %s.%s declara tipo inexistente %s (linha %d)", fm.Name, info.Name, n.Name, fm.Type, fm.Line)
					}
				}
			}
		}
	}
	return nil
}

// resolve ancora SELF_TYPE na classe corrente: dentro de C,
// SELF_TYPE resolve para C. Qualquer outro nome passa intacto.
func resolve(typ, current string) string {
	if typ == "SELF_TYPE" {
		return current
	}
	return typ
}

// conformsTo diz se um valor de tipo estático sub pode ir onde se espera
// o tipo declarado sup, ambos lidos na classe current.
func (t *Table) conformsTo(sub, sup, current string) bool {
	if sup == "SELF_TYPE" {
		return sub == "SELF_TYPE"
	}
	if sub == "SELF_TYPE" {
		return t.Conforms(current, sup)
	}
	return t.Conforms(sub, sup)
}
