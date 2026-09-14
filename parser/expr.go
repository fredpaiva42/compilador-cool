package parser

import (
	"cool/ast"
	"cool/token"
)

func (p *Parser) parseExpr() (ast.Expr, error) {
	cur := p.peek()
	switch cur.Type {
	case token.NEW:
		p.next() // consome o "new"
		typeTok, err := p.expect(token.TYPEID)
		if err != nil {
			return nil, err
		}
		return &ast.New{Type: typeTok.Literal, Line: cur.Line, Column: cur.Column}, nil
	case token.ISVOID:
		p.next()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.IsVoid{Expr: e, Line: cur.Line, Column: cur.Column}, nil
	default:
		return p.parseDispatch()
	}
}
