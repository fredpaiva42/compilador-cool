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
			return nil, err
		}

		prog.Classes = append(prog.Classes, cls)
	}

	if len(prog.Classes) == 0 {
		cur := p.peek()
		return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "programa vazio: esperava ao menos uma classe"}
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

	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}

	if _, err := p.expect(token.SEMI); err != nil {
		return nil, err
	}

	return &ast.Class{
		Name:     nameTok.Literal,
		Parent:   parent,
		Features: nil,
		Line:     classTok.Line,
		Column:   classTok.Column,
	}, nil
}
