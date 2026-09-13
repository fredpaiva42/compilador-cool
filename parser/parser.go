package parser

import (
	"cool/token"
	"fmt"
)

type ParseError struct {
	Line   int
	Column int
	Msg    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Msg)
}

type Parser struct {
	toks []token.Token
	pos  int
}

func New(toks []token.Token) *Parser {
	return &Parser{toks: toks}
}

func (p *Parser) peek() token.Token {
	if p.pos < len(p.toks) {
		return p.toks[p.pos]
	}

	if len(p.toks) > 0 {
		return p.toks[len(p.toks)-1]
	}

	return token.Token{Type: token.EOF}
}

func (p *Parser) next() token.Token {
	t := p.peek()
	if p.pos < len(p.toks) {
		p.pos++
	}

	return t
}

func (p *Parser) check(t token.Type) bool {
	return p.peek().Type == t
}

func (p *Parser) atEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) expect(t token.Type) (token.Token, error) {
	cur := p.peek()
	if cur.Type != t {
		return cur, &ParseError{
			Line:   cur.Line,
			Column: cur.Column,
			Msg:    fmt.Sprintf("esperado %s, encontrado %s (%q)", t, cur.Type, cur.Literal),
		}
	}
	p.pos++
	return cur, nil
}
