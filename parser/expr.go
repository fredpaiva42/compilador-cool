package parser

import (
	"cool/ast"
)

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseAssign()
}
