package parser

import (
	"cool/ast"
	"cool/token"
	"fmt"
)

func (p *Parser) parseFormal() (*ast.Formal, error) {
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

	return &ast.Formal{
		Name:   nameTok.Literal,
		Type:   typeTok.Literal,
		Line:   nameTok.Line,
		Column: nameTok.Column,
	}, nil
}

func (p *Parser) parseAttributeRest(nameTok token.Token) (*ast.Attribute, error) {
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}

	typeTok, err := p.expect(token.TYPEID)
	if err != nil {
		return nil, err
	}

	attr := &ast.Attribute{
		Name:   nameTok.Literal,
		Type:   typeTok.Literal,
		Line:   nameTok.Line,
		Column: nameTok.Column,
	}

	if p.check(token.ASSIGN) {
		p.next()
		init, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		attr.Init = init
		attr.HasInit = true
	}

	if _, err := p.expect(token.SEMI); err != nil {
		return nil, err
	}

	return attr, nil
}

func (p *Parser) parseMethodRest(nameTok token.Token) (*ast.Method, error) {
	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}

	var formals []*ast.Formal
	if !p.check(token.RPAREN) {
		for {
			f, err := p.parseFormal()
			if err != nil {
				return nil, err
			}
			formals = append(formals, f)
			if p.check(token.COMMA) {
				p.next()
				continue
			}
			break
		}
	}

	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}

	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}

	retTok, err := p.expect(token.TYPEID)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}

	body, err := p.parseExpr()
	if err != nil {

		if !p.syncMethodBody() {
			return nil, err
		}
		p.errs = append(p.errs, err)
		p.next()
		if _, serr := p.expect(token.SEMI); serr != nil {
			p.errs = append(p.errs, serr)
		}
		return &ast.Method{
			Name:       nameTok.Literal,
			Formals:    formals,
			ReturnType: retTok.Literal,
			Body:       nil,
			Line:       nameTok.Line,
			Column:     nameTok.Column,
		}, nil
	}

	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}

	if _, err := p.expect(token.SEMI); err != nil {
		return nil, err
	}

	return &ast.Method{
		Name:       nameTok.Literal,
		Formals:    formals,
		ReturnType: retTok.Literal,
		Body:       body,
		Line:       nameTok.Line,
		Column:     nameTok.Column,
	}, nil
}

func (p *Parser) parseFeature() (ast.Feature, error) {
	nameTok, err := p.expect(token.OBJECTID)
	if err != nil {
		return nil, err
	}

	if p.check(token.LPAREN) {
		return p.parseMethodRest(nameTok)
	}

	if p.check(token.COLON) {
		return p.parseAttributeRest(nameTok)
	}

	cur := p.peek()
	return nil, &ParseError{
		Line:   cur.Line,
		Column: cur.Column,
		Msg:    fmt.Sprintf("após nome de feature esperava '(' ou ':', encontrei %s (%q)", cur.Type, cur.Literal),
	}
}
