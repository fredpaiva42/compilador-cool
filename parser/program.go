package parser

import (
	"cool/ast"
	"cool/token"
)

func (p *Parser) ParseProgram() (*ast.Program, error) {
	prog := &ast.Program{}
	for !p.atEnd() {
		cls, err := p.parseClass()
		if err != nil {
			p.errs = append(p.errs, err)
			p.syncTo(token.CLASS)
			continue
		}
		prog.Classes = append(prog.Classes, cls)
	}
	if len(prog.Classes) == 0 && len(p.errs) == 0 {
		cur := p.peek()
		return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "programa vazio: esperava ao menos uma classe"}
	}
	if len(p.errs) > 0 {
		return prog, &ParseErrors{Errs: p.errs}
	}
	return prog, nil
}

func (p *Parser) parseClass() (*ast.Class, error) {
	classTok, err := p.expect(token.CLASS)
	if err != nil {
		return nil, err
	}
	nameTok, err := p.expect(token.TYPEID)
	if err != nil {
		return nil, err
	}
	parent := ""
	if p.check(token.INHERITS) {
		p.next()
		parentTok, err := p.expect(token.TYPEID)
		if err != nil {
			return nil, err
		}
		parent = parentTok.Literal
	}
	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}

	var feats []ast.Feature
	for !p.check(token.RBRACE) {
		if p.atEnd() {
			cur := p.peek()
			return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "fim de arquivo dentro de classe: faltou '}'"}
		}
		f, err := p.parseFeature()
		if err != nil {
			p.errs = append(p.errs, err)
			p.syncTo(token.SEMI, token.RBRACE)
			if p.check(token.SEMI) {
				p.next()
			}
			continue
		}
		feats = append(feats, f)
	}
	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.SEMI); err != nil {
		return nil, err
	}
	return &ast.Class{
		Name:     nameTok.Literal,
		Parent:   parent,
		Features: feats,
		Line:     classTok.Line,
		Column:   classTok.Column,
	}, nil
}
