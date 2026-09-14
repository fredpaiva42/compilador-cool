package parser

import (
	"cool/ast"
	"cool/token"
)

func (p *Parser) parseAssign() (ast.Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	if p.check(token.ASSIGN) {
		p.next()
		obj, ok := left.(*ast.Object)
		if !ok {
			l, c := recvPos(left)
			return nil, &ParseError{Line: l, Column: c, Msg: "alvo de atribuição deve ser identificador"}
		}
		right, err := p.parseAssign()
		if err != nil {
			return nil, err
		}
		return &ast.Assign{Target: obj.Name, Value: right, Line: obj.Line, Column: obj.Column}, nil
	}
	return left, nil
}

func (p *Parser) parseNot() (ast.Expr, error) {
	if p.check(token.NOT) {
		nTok := p.next()
		e, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &ast.Not{Expr: e, Line: nTok.Line, Column: nTok.Column}, nil
	}
	return p.parseComparison()
}

func (p *Parser) parseComparison() (ast.Expr, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	if p.check(token.LT) || p.check(token.LE) || p.check(token.EQ) {
		opTok := p.next()
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		if p.check(token.LT) || p.check(token.LE) || p.check(token.EQ) {
			cur := p.peek()
			return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "comparações não associam: use parênteses"}
		}
		l, c := recvPos(left)
		return &ast.BinOp{Op: opTok.Literal, Left: left, Right: right, Line: l, Column: c}, nil
	}
	return left, nil
}

func (p *Parser) parseAdd() (ast.Expr, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.check(token.PLUS) || p.check(token.MINUS) {
		opTok := p.next()
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		l, c := recvPos(left)
		left = &ast.BinOp{Op: opTok.Literal, Left: left, Right: right, Line: l, Column: c}
	}
	return left, nil
}

func (p *Parser) parseMul() (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.check(token.MULT) || p.check(token.DIV) {
		opTok := p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		l, c := recvPos(left)
		left = &ast.BinOp{Op: opTok.Literal, Left: left, Right: right, Line: l, Column: c}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	cur := p.peek()
	switch cur.Type {
	case token.NEG:
		p.next()
		e, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.Neg{Expr: e, Line: cur.Line, Column: cur.Column}, nil
	case token.ISVOID:
		p.next()
		e, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.IsVoid{Expr: e, Line: cur.Line, Column: cur.Column}, nil
	case token.IF:
		return p.parseIf()
	case token.WHILE:
		return p.parseWhile()
	case token.LBRACE:
		return p.parseBlock()
	case token.LET:
		return p.parseLet()
	case token.CASE:
		return p.parseCase()
	default:
		return p.parseDispatch()
	}
}
