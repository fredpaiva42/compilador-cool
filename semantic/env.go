package semantic

import (
	"cool/ast"
	"fmt"
)

type MethodSig struct {
	Name       string
	ParamNames []string
	ParamTypes []string
	Return     string
	Owner      string
	Line       int
}

type AttrInfo struct {
	Name  string
	Type  string
	Owner string
	Line  int
}

func isBuiltin(name string) bool {
	switch name {
	case "Object", "IO", "Int", "String", "Bool":
		return true
	}
	return false
}

func (t *Table) BuildEnvs() error {
	// 1. Proibido herdar de Int/String/Bool.
	for _, info := range t.Classes {
		if info.Node == nil {
			continue // básica, sem declaração
		}
		switch info.Parent {
		case "Int", "String", "Bool":
			return fmt.Errorf("classe %s não pode herdar de %s (linha %d)", info.Name, info.Parent, info.Node.Line)
		}
	}

	// 2. Métodos das básicas (o resto elas herdam de Object pelo fluxo normal).
	builtinMethods := map[string][]*MethodSig{
		"Object": {
			{Name: "abort", Return: "Object", Owner: "Object"},
			{Name: "type_name", Return: "String", Owner: "Object"},
			{Name: "copy", Return: "SELF_TYPE", Owner: "Object"},
		},
		"IO": {
			{Name: "out_string", ParamNames: []string{"x"}, ParamTypes: []string{"String"}, Return: "SELF_TYPE", Owner: "IO"},
			{Name: "out_int", ParamNames: []string{"x"}, ParamTypes: []string{"Int"}, Return: "SELF_TYPE", Owner: "IO"},
			{Name: "in_string", Return: "String", Owner: "IO"},
			{Name: "in_int", Return: "Int", Owner: "IO"},
		},
		"String": {
			{Name: "length", Return: "Int", Owner: "String"},
			{Name: "concat", ParamNames: []string{"s"}, ParamTypes: []string{"String"}, Return: "String", Owner: "String"},
			{Name: "substr", ParamNames: []string{"i", "l"}, ParamTypes: []string{"Int", "Int"}, Return: "String", Owner: "String"},
		},
	}
	t.Methods = map[string]map[string]*MethodSig{}
	t.Attrs = map[string]map[string]*AttrInfo{}
	for cls, sigs := range builtinMethods {
		m := map[string]*MethodSig{}
		for _, s := range sigs {
			m[s.Name] = s
		}
		t.Methods[cls] = m
		t.Attrs[cls] = map[string]*AttrInfo{}
	}

	// Básicas com métodos próprios também herdam os de Object:
	// abort/type_name/copy existem em toda classe.
	for _, cls := range []string{"IO", "String"} {
		for k, v := range t.Methods["Object"] {
			if _, ok := t.Methods[cls][k]; !ok {
				t.Methods[cls][k] = v
			}
		}
	}

	// 3. Usuárias em ordem pai-antes-do-filho, copiando os herdados.
	for _, name := range t.order() {
		if _, done := t.Methods[name]; done {
			continue // básica com métodos próprios já registrada
		}

		info := t.Classes[name]
		methods := map[string]*MethodSig{}
		attrs := map[string]*AttrInfo{}
		parent := info.Parent
		if parent == "" && name != "Object" {
			parent = "Object" // herança implícita
		}

		if parent != "" {
			for k, v := range t.Methods[parent] {
				methods[k] = v
			}
			for k, v := range t.Attrs[parent] {
				attrs[k] = v
			}
		}

		ownMethods := map[string]bool{}
		ownAttrs := map[string]bool{}
		if info.Node != nil {
			for _, f := range info.Node.Features {
				switch n := f.(type) {
				case *ast.Attribute:
					if ownAttrs[n.Name] {
						return fmt.Errorf("atributo %s duplicado na classe %s (linha %d)", n.Name, name, n.Line)
					}
					if _, inherited := attrs[n.Name]; inherited {
						return fmt.Errorf("atributo %s de %s redefine herdado (linha %d)", n.Name, name, n.Line)
					}
					ownAttrs[n.Name] = true
					attrs[n.Name] = &AttrInfo{Name: n.Name, Type: n.Type, Owner: name, Line: n.Line}
				case *ast.Method:
					if ownMethods[n.Name] {
						return fmt.Errorf("método %s duplicado na classe %s (linha %d)", n.Name, name, n.Line)
					}
					ownMethods[n.Name] = true
					sig := &MethodSig{Name: n.Name, Return: n.ReturnType, Owner: name, Line: n.Line}
					for _, fm := range n.Formals {
						sig.ParamNames = append(sig.ParamNames, fm.Name)
						sig.ParamTypes = append(sig.ParamTypes, fm.Type)
					}

					if err := t.checkOverride(name, n.Name, sig); err != nil {
						return err
					}
					methods[n.Name] = sig
				}
			}
		}
		t.Methods[name] = methods
		t.Attrs[name] = attrs
	}
	return nil
}

func (t *Table) order() []string {
	visited := map[string]bool{}
	var out []string
	var visit func(n string)
	visit = func(n string) {
		if visited[n] {
			return
		}
		visited[n] = true
		info := t.Classes[n]
		if info.Parent != "" {
			visit(info.Parent)
		} else if n != "Object" {
			visit("Object")
		}
		out = append(out, n)
	}
	for name := range t.Classes {
		visit(name)
	}
	return out
}
