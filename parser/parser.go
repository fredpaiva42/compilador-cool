package parser

import (
	"cool/token"
	"fmt"
	"strings"
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
	errs []error
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

type ParseErrors struct{ Errs []error }

func (e *ParseErrors) Error() string {
	msgs := make([]string, len(e.Errs))
	for i, err := range e.Errs {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}

func (p *Parser) syncTo(sync ...token.Type) {
loop:
	for !p.atEnd() {
		for _, t := range sync {
			if p.check(t) {
				break loop
			}
		}
		p.next()
	}
}

func (p *Parser) syncMethodBody() bool {
	depth := 0
	for !p.atEnd() {
		if p.check(token.RBRACE) && depth == 0 {
			return true
		}
		if p.check(token.LBRACE) {
			depth++
		} else if p.check(token.RBRACE) {
			depth--
		}
		p.next()
	}
	return false
}
