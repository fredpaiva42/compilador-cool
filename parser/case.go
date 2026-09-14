package parser

import (
	"cool/ast"
	"cool/token"
)

func (p *Parser) parseCase() (ast.Expr, error) {
	cTok, err := p.expect(token.CASE)
	if err != nil {
		return nil, err
	}
	subj, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.OF); err != nil {
		return nil, err
	}
	var branches []*ast.CaseBranch
	for !p.check(token.ESAC) {
		if p.atEnd() {
			cur := p.peek()
			return nil, &ParseError{Line: cur.Line, Column: cur.Column, Msg: "fim de arquivo dentro de case: faltou 'esac'"}
		}
		b, err := p.parseCaseBranch()
		if err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	if len(branches) == 0 {
		return nil, &ParseError{Line: cTok.Line, Column: cTok.Column, Msg: "case sem ramos: esperava ao menos um 'nome : Tipo => expr;'"}
	}
	if _, err := p.expect(token.ESAC); err != nil {
		return nil, err
	}
	return &ast.Case{Subject: subj, Branches: branches, Line: cTok.Line, Column: cTok.Column}, nil
}

func (p *Parser) parseCaseBranch() (*ast.CaseBranch, error) {
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
	if _, err := p.expect(token.DARROW); err != nil {
		return nil, err
	}
	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.SEMI); err != nil {
		return nil, err
	}
	return &ast.CaseBranch{Name: nameTok.Literal, Type: typeTok.Literal, Body: body, Line: nameTok.Line, Column: nameTok.Column}, nil
}
