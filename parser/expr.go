package parser

import (
	"cool/ast"
)

func (p *Parser) parseExpr() (ast.Expr, error) {
	cur := p.peek()
	return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "parseExpr ainda não implementado"}
}
