package semantic

import (
	"fmt"

	"cool/ast"
)

// typeEnv é o ambiente de checagem de uma expressão (§12.1, parte O + C):
// as variáveis visíveis e a classe onde o código está.
type typeEnv struct {
	tab     *Table
	current string            // classe corrente (ancora do SELF_TYPE)
	vars    map[string]string // OBJECTID → tipo declarado (pode ser SELF_TYPE)
}

// checkExpr infere o tipo estático de qualquer expressão.
// Devolve o nome do tipo ("Int", "Cachorro", "SELF_TYPE"...).
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
	default:
		return "", fmt.Errorf("expressão ainda sem checagem (linha %d)", exprLine(x))
	}
}

// checkNew confere "new T": SELF_TYPE sempre vale; o resto precisa existir.
func (e *typeEnv) checkNew(n *ast.New) (string, error) {
	if n.Type == "SELF_TYPE" {
		return "SELF_TYPE", nil
	}
	if _, ok := e.tab.Classes[n.Type]; !ok {
		return "", fmt.Errorf("new de tipo inexistente %s (linha %d)", n.Type, n.Line)
	}
	return n.Type, nil
}

// checkBinOp confere operadores infixos (§7.12).
// Os tipos são resolvidos (SELF_TYPE → corrente) antes de comparar.
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

// checkObject resolve um uso de variável: self vale SELF_TYPE,
// o resto é procurado no ambiente (formais + atributos entram no Passo 7).
func (e *typeEnv) checkObject(n *ast.Object) (string, error) {
	if n.Name == "self" {
		return "SELF_TYPE", nil
	}
	if t, ok := e.vars[n.Name]; ok {
		return t, nil
	}
	return "", fmt.Errorf("identificador %s não declarado (linha %d)", n.Name, n.Line)
}

// exprLine extrai a linha de qualquer nó (para erros).
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
	default:
		return 0
	}
}
