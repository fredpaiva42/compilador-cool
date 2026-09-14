package parser

import (
	"strings"

	"cool/ast"
	"cool/token"
)

func (p *Parser) parsePrimary() (ast.Expr, error) {
	cur := p.peek()
	switch cur.Type {
	case token.INT_CONST:
		p.next()
		return &ast.IntConst{Value: cur.Literal, Line: cur.Line, Column: cur.Column}, nil
	case token.STR_CONST:
		p.next()
		return &ast.StringConst{Value: cur.Literal, Line: cur.Line, Column: cur.Column}, nil
	case token.BOOL_CONST:
		p.next()
		return &ast.BoolConst{Value: strings.EqualFold(cur.Literal, "true"), Line: cur.Line, Column: cur.Column}, nil
	case token.OBJECTID:
		p.next()
		return &ast.Object{Name: cur.Literal, Line: cur.Line, Column: cur.Column}, nil
	case token.LPAREN:
		p.next() // consome "("
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.RPAREN); err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "esperava expressão, encontrei " + cur.Type.String() + " (" + cur.Literal + ")"}
	}
}
