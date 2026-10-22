package semantic

import (
	"fmt"

	"cool/ast"
)

// CheckProgram executa a análise semântica na ordem em que as informações
// ficam disponíveis: esqueleto da hierarquia, ambientes com os herdados,
// tipos declarados, corpos e inits das classes, e o contrato de Main.main.
func CheckProgram(prog *ast.Program) error {
	tab, err := Build(prog)
	if err != nil {
		return err
	}
	if err := tab.BuildEnvs(); err != nil {
		return err
	}
	if err := tab.checkDeclaredTypes(); err != nil {
		return err
	}
	for _, name := range tab.order() {
		info := tab.Classes[name]
		if info.Node == nil {
			continue // básica: sem corpo a checar
		}
		if err := tab.checkClass(info); err != nil {
			return err
		}
	}
	return tab.checkMain()
}

func (t *Table) classEnv(className string) *typeEnv {
	vars := map[string]string{}
	for name, a := range t.Attrs[className] {
		vars[name] = a.Type
	}
	vars["self"] = "SELF_TYPE"
	return &typeEnv{tab: t, current: className, vars: vars}
}

// Sem init vale o valor padrão do tipo, então não há o que checar.
func (t *Table) checkAttribute(className string, a *ast.Attribute) error {
	if !a.HasInit {
		return nil
	}
	e := t.classEnv(className)
	it, err := e.checkExpr(a.Init)
	if err != nil {
		return err
	}
	if !t.conformsTo(it, a.Type, className) {
		return fmt.Errorf("init de %s.%s: %s não conforma com %s (linha %d)",
			className, a.Name, it, a.Type, a.Line)
	}
	return nil
}

// Formal com nome de atributo esconde o atributo dentro do método.
func (t *Table) checkMethod(className string, m *ast.Method) error {
	e := t.classEnv(className)
	seen := map[string]bool{}
	for _, fm := range m.Formals {
		if seen[fm.Name] {
			return fmt.Errorf("formal %s duplicado em %s.%s (linha %d)", fm.Name, className, m.Name, fm.Line)
		}
		seen[fm.Name] = true
		e.vars[fm.Name] = fm.Type
	}
	bt, err := e.checkExpr(m.Body)
	if err != nil {
		return err
	}
	if !t.conformsTo(bt, m.ReturnType, className) {
		return fmt.Errorf("corpo de %s.%s: %s não conforma com %s (linha %d)",
			className, m.Name, bt, m.ReturnType, m.Line)
	}
	return nil
}

func (t *Table) checkClass(info *ClassInfo) error {
	for _, f := range info.Node.Features {
		switch n := f.(type) {
		case *ast.Attribute:
			if err := t.checkAttribute(info.Name, n); err != nil {
				return err
			}
		case *ast.Method:
			if err := t.checkMethod(info.Name, n); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkMain exige main próprio, sem formais, e não herdado.
func (t *Table) checkMain() error {
	main, ok := t.Classes["Main"]
	if !ok || main.Node == nil {
		return fmt.Errorf("programa sem classe Main")
	}
	sig, ok := t.Methods["Main"]["main"]
	if !ok || sig.Owner != "Main" {
		return fmt.Errorf("Main sem método main próprio")
	}
	if len(sig.ParamTypes) != 0 {
		return fmt.Errorf("Main.main não pode ter formais")
	}
	return nil
}
