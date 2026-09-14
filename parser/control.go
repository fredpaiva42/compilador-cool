package parser

import (
	"cool/ast"
	"cool/token"
)

func (p *Parser) parseIf() (ast.Expr, error) {
	ifTok, err := p.expect(token.IF)
	if err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.THEN); err != nil {
		return nil, err
	}
	then, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.ELSE); err != nil {
		return nil, err
	}
	els, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.FI); err != nil {
		return nil, err
	}
	return &ast.If{Cond: cond, Then: then, Else: els, Line: ifTok.Line, Column: ifTok.Column}, nil
}

func (p *Parser) parseWhile() (ast.Expr, error) {
	wTok, err := p.expect(token.WHILE)
	if err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.LOOP); err != nil {
		return nil, err
	}
	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.POOL); err != nil {
		return nil, err
	}
	return &ast.While{Cond: cond, Body: body, Line: wTok.Line, Column: wTok.Column}, nil
}

func (p *Parser) parseBlock() (ast.Expr, error) {
	lTok, err := p.expect(token.LBRACE)
	if err != nil {
		return nil, err
	}
	var exprs []ast.Expr
	for !p.check(token.RBRACE) {
		if p.atEnd() {
			cur := p.peek()
			return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "fim de arquivo dentro de bloco: faltou '}'"}
		}
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.SEMI); err != nil {
			return nil, err
		}
		exprs = append(exprs, e)
	}
	if len(exprs) == 0 {
		return nil, &ParseError{Line: lTok.Line, Column: lTok.Column, Msg: "bloco vazio: esperava ao menos uma expressão"}
	}
	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	return &ast.Block{Exprs: exprs, Line: lTok.Line, Column: lTok.Column}, nil
}

func (p *Parser) parseLetBinding() (*ast.LetBinding, error) {
	nameTok, err := p.expect(token.OBJECTID)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	typeTok, err := p.expect(token.TYPEID)
	if err != nil {
		return nil, err
	}
	b := &ast.LetBinding{Name: nameTok.Literal, Type: typeTok.Literal, Line: nameTok.Line, Column: nameTok.Column}
	if p.check(token.ASSIGN) {
		p.next()
		init, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		b.Init = init
		b.HasInit = true
	}
	return b, nil
}

func (p *Parser) parseLet() (ast.Expr, error) {
	lTok, err := p.expect(token.LET)
	if err != nil {
		return nil, err
	}
	var bindings []*ast.LetBinding
	for {
		b, err := p.parseLetBinding()
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, b)
		if p.check(token.COMMA) {
			p.next()
			continue
		}
		break
	}
	if _, err := p.expect(token.IN); err != nil {
		return nil, err
	}
	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.Let{Bindings: bindings, Body: body, Line: lTok.Line, Column: lTok.Column}, nil
}
