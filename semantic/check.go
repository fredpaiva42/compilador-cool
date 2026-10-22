package semantic

import (
	"fmt"

	"cool/ast"
)

type typeEnv struct {
	tab     *Table
	current string
	vars    map[string]string // valores podem ser "SELF_TYPE"
}

// checkExpr devolve o tipo estático de x como nome ("Int", "Cachorro", "SELF_TYPE").
func (e *typeEnv) checkExpr(x ast.Expr) (string, error) {
	switch n := x.(type) {
	case *ast.IntConst:
		return "Int", nil
	case *ast.StringConst:
		return "String", nil
	case *ast.BoolConst:
		return "Bool", nil
	case *ast.Object:
		return e.checkObject(n)
	case *ast.New:
		return e.checkNew(n)
	case *ast.IsVoid:
		if _, err := e.checkExpr(n.Expr); err != nil {
			return "", err
		}
		return "Bool", nil
	case *ast.Neg:
		t, err := e.checkExpr(n.Expr)
		if err != nil {
			return "", err
		}
		if resolve(t, e.current) != "Int" {
			return "", fmt.Errorf("~ exige Int, encontrei %s (linha %d)", t, n.Line)
		}
		return "Int", nil
	case *ast.Not:
		t, err := e.checkExpr(n.Expr)
		if err != nil {
			return "", err
		}
		if resolve(t, e.current) != "Bool" {
			return "", fmt.Errorf("not exige Bool, encontrei %s (linha %d)", t, n.Line)
		}
		return "Bool", nil
	case *ast.BinOp:
		return e.checkBinOp(n)
	case *ast.Dispatch:
		return e.checkDispatch(n)
	case *ast.StaticDispatch:
		return e.checkStaticDispatch(n)
	case *ast.Assign:
		return e.checkAssign(n)
	case *ast.Block:
		return e.checkBlock(n)
	case *ast.If:
		return e.checkIf(n)
	case *ast.While:
		return e.checkWhile(n)
	case *ast.Let:
		return e.checkLet(n)
	case *ast.Case:
		return e.checkCase(n)
	default:
		return "", fmt.Errorf("expressão não tratada pelo verificador (linha %d)", exprLine(x))
	}
}

func (e *typeEnv) checkNew(n *ast.New) (string, error) {
	if n.Type == "SELF_TYPE" {
		return "SELF_TYPE", nil
	}
	if _, ok := e.tab.Classes[n.Type]; !ok {
		return "", fmt.Errorf("new de tipo inexistente %s (linha %d)", n.Type, n.Line)
	}
	return n.Type, nil
}

// checkBinOp resolve SELF_TYPE antes de comparar os operandos.
func (e *typeEnv) checkBinOp(n *ast.BinOp) (string, error) {
	l, err := e.checkExpr(n.Left)
	if err != nil {
		return "", err
	}
	r, err := e.checkExpr(n.Right)
	if err != nil {
		return "", err
	}
	ls, rs := resolve(l, e.current), resolve(r, e.current)
	switch n.Op {
	case "+", "-", "*", "/":
		if ls != "Int" || rs != "Int" {
			return "", fmt.Errorf("operador %s exige Int dos dois lados (linha %d)", n.Op, n.Line)
		}
		return "Int", nil
	case "<", "<=":
		if ls != "Int" || rs != "Int" {
			return "", fmt.Errorf("operador %s exige Int dos dois lados (linha %d)", n.Op, n.Line)
		}
		return "Bool", nil
	case "=":
		if ls == "Int" || ls == "String" || ls == "Bool" ||
			rs == "Int" || rs == "String" || rs == "Bool" {
			if ls != rs {
				return "", fmt.Errorf("= entre %s e %s: básicas só comparam com o mesmo tipo (linha %d)", ls, rs, n.Line)
			}
		}
		return "Bool", nil
	default:
		return "", fmt.Errorf("operador desconhecido %s (linha %d)", n.Op, n.Line)
	}
}

func (e *typeEnv) checkObject(n *ast.Object) (string, error) {
	if n.Name == "self" {
		return "SELF_TYPE", nil
	}
	if t, ok := e.vars[n.Name]; ok {
		return t, nil
	}
	return "", fmt.Errorf("identificador %s não declarado (linha %d)", n.Name, n.Line)
}

func exprLine(x ast.Expr) int {
	switch n := x.(type) {
	case *ast.IntConst:
		return n.Line
	case *ast.StringConst:
		return n.Line
	case *ast.BoolConst:
		return n.Line
	case *ast.Object:
		return n.Line
	case *ast.New:
		return n.Line
	case *ast.IsVoid:
		return n.Line
	case *ast.Neg:
		return n.Line
	case *ast.Not:
		return n.Line
	case *ast.BinOp:
		return n.Line
	case *ast.Dispatch:
		return n.Line
	case *ast.StaticDispatch:
		return n.Line
	case *ast.Assign:
		return n.Line
	case *ast.Block:
		return n.Line
	case *ast.If:
		return n.Line
	case *ast.While:
		return n.Line
	case *ast.Let:
		return n.Line
	case *ast.Case:
		return n.Line
	default:
		return 0
	}
}

// lookupMethod consulta direto: o BuildEnvs já copiou os herdados.
func (e *typeEnv) lookupMethod(cls, name string) (*MethodSig, bool) {
	m, ok := e.tab.Methods[cls][name]
	return m, ok
}

// Receptor nil é self implícito, que entra como SELF_TYPE.
func (e *typeEnv) checkDispatch(n *ast.Dispatch) (string, error) {
	recvType := "SELF_TYPE"
	if n.Receiver != nil {
		var err error
		recvType, err = e.checkExpr(n.Receiver)
		if err != nil {
			return "", err
		}
	}
	lookup := resolve(recvType, e.current)
	sig, ok := e.lookupMethod(lookup, n.Method)
	if !ok {
		return "", fmt.Errorf("método %s inexistente em %s (linha %d)", n.Method, lookup, n.Line)
	}
	if len(n.Args) != len(sig.ParamTypes) {
		return "", fmt.Errorf("método %s espera %d argumento(s), recebeu %d (linha %d)",
			n.Method, len(sig.ParamTypes), len(n.Args), n.Line)
	}
	for i, a := range n.Args {
		at, err := e.checkExpr(a)
		if err != nil {
			return "", err
		}
		if !e.tab.conformsTo(at, sig.ParamTypes[i], e.current) {
			return "", fmt.Errorf("argumento %d de %s: %s não conforma com %s (linha %d)",
				i+1, n.Method, at, sig.ParamTypes[i], n.Line)
		}
	}
	if sig.Return == "SELF_TYPE" {
		return recvType, nil
	}
	return sig.Return, nil
}

func (e *typeEnv) checkStaticDispatch(n *ast.StaticDispatch) (string, error) {
	recvType, err := e.checkExpr(n.Receiver)
	if err != nil {
		return "", err
	}
	if n.StaticType != "SELF_TYPE" {
		if _, ok := e.tab.Classes[n.StaticType]; !ok {
			return "", fmt.Errorf("tipo %s inexistente no dispatch estático (linha %d)", n.StaticType, n.Line)
		}
	}
	if !e.tab.conformsTo(recvType, n.StaticType, e.current) {
		return "", fmt.Errorf("receptor %s não conforma com %s no dispatch estático (linha %d)",
			recvType, n.StaticType, n.Line)
	}
	lookup := resolve(n.StaticType, e.current)
	sig, ok := e.lookupMethod(lookup, n.Method)
	if !ok {
		return "", fmt.Errorf("método %s inexistente em %s (linha %d)", n.Method, lookup, n.Line)
	}
	if len(n.Args) != len(sig.ParamTypes) {
		return "", fmt.Errorf("método %s espera %d argumento(s), recebeu %d (linha %d)",
			n.Method, len(sig.ParamTypes), len(n.Args), n.Line)
	}
	for i, a := range n.Args {
		at, err := e.checkExpr(a)
		if err != nil {
			return "", err
		}
		if !e.tab.conformsTo(at, sig.ParamTypes[i], e.current) {
			return "", fmt.Errorf("argumento %d de %s: %s não conforma com %s (linha %d)",
				i+1, n.Method, at, sig.ParamTypes[i], n.Line)
		}
	}
	if sig.Return == "SELF_TYPE" {
		return recvType, nil
	}
	return sig.Return, nil
}

// checkAssign segue a regra [ASSIGN] do manual: o valor precisa conformar com
// o alvo, e o tipo resultante é o do valor, não o do alvo.
func (e *typeEnv) checkAssign(n *ast.Assign) (string, error) {
	if n.Target == "self" {
		return "", fmt.Errorf("atribuição para self é proibida (linha %d)", n.Line)
	}
	decl, ok := e.vars[n.Target]
	if !ok {
		return "", fmt.Errorf("atribuição a %s não declarado (linha %d)", n.Target, n.Line)
	}
	vt, err := e.checkExpr(n.Value)
	if err != nil {
		return "", err
	}
	if !e.tab.conformsTo(vt, decl, e.current) {
		return "", fmt.Errorf("atribuição: %s não conforma com %s (linha %d)", vt, decl, n.Line)
	}
	return vt, nil
}

// O tipo do bloco é o do último item.
func (e *typeEnv) checkBlock(n *ast.Block) (string, error) {
	var t string
	for _, x := range n.Exprs {
		var err error
		t, err = e.checkExpr(x)
		if err != nil {
			return "", err
		}
	}
	return t, nil
}

// O tipo do if é o join dos dois ramos; a condição precisa ser Bool.
func (e *typeEnv) checkIf(n *ast.If) (string, error) {
	c, err := e.checkExpr(n.Cond)
	if err != nil {
		return "", err
	}
	if resolve(c, e.current) != "Bool" {
		return "", fmt.Errorf("condição do if exige Bool, encontrei %s (linha %d)", c, n.Line)
	}
	t, err := e.checkExpr(n.Then)
	if err != nil {
		return "", err
	}
	f, err := e.checkExpr(n.Else)
	if err != nil {
		return "", err
	}
	return e.tab.Join(resolve(t, e.current), resolve(f, e.current)), nil
}

// while tem tipo Object; a condição precisa ser Bool.
func (e *typeEnv) checkWhile(n *ast.While) (string, error) {
	c, err := e.checkExpr(n.Cond)
	if err != nil {
		return "", err
	}
	if resolve(c, e.current) != "Bool" {
		return "", fmt.Errorf("condição do while exige Bool, encontrei %s (linha %d)", c, n.Line)
	}
	if _, err := e.checkExpr(n.Body); err != nil {
		return "", err
	}
	return "Object", nil
}

func (e *typeEnv) checkLet(n *ast.Let) (string, error) {
	vars := map[string]string{}
	for k, v := range e.vars {
		vars[k] = v // cópia: o escopo do let morre aqui dentro
	}
	inner := &typeEnv{tab: e.tab, current: e.current, vars: vars}
	for _, b := range n.Bindings {
		if b.Type != "SELF_TYPE" {
			if _, ok := e.tab.Classes[b.Type]; !ok {
				return "", fmt.Errorf("let %s declara tipo inexistente %s (linha %d)", b.Name, b.Type, b.Line)
			}
		}
		if b.HasInit {
			it, err := inner.checkExpr(b.Init)
			if err != nil {
				return "", err
			}
			if !e.tab.conformsTo(it, b.Type, e.current) {
				return "", fmt.Errorf("init de %s: %s não conforma com %s (linha %d)", b.Name, it, b.Type, b.Line)
			}
		}
		vars[b.Name] = b.Type
	}
	return inner.checkExpr(n.Body)
}

func (e *typeEnv) checkCase(n *ast.Case) (string, error) {
	if _, err := e.checkExpr(n.Subject); err != nil {
		return "", err
	}
	if len(n.Branches) == 0 {
		return "", fmt.Errorf("case sem ramos (linha %d)", n.Line)
	}
	seen := map[string]bool{}
	var types []string
	for _, b := range n.Branches {
		if b.Type == "SELF_TYPE" {
			return "", fmt.Errorf("ramo do case não pode ser SELF_TYPE (linha %d)", b.Line)
		}
		if _, ok := e.tab.Classes[b.Type]; !ok {
			return "", fmt.Errorf("ramo declara tipo inexistente %s (linha %d)", b.Type, b.Line)
		}
		if seen[b.Type] {
			return "", fmt.Errorf("tipo %s repetido nos ramos do case (linha %d)", b.Type, b.Line)
		}
		seen[b.Type] = true
		vars := map[string]string{}
		for k, v := range e.vars {
			vars[k] = v
		}
		vars[b.Name] = b.Type
		inner := &typeEnv{tab: e.tab, current: e.current, vars: vars}
		bt, err := inner.checkExpr(b.Body)
		if err != nil {
			return "", err
		}
		types = append(types, resolve(bt, e.current))
	}
	return e.tab.JoinAll(types), nil
}
