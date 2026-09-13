package parser

import (
	"cool/token"
	"testing"
)

func TestCursor(t *testing.T) {
	toks := []token.Token{
		{Type: token.CLASS, Literal: "class", Line: 1, Column: 1},
		{Type: token.TYPEID, Literal: "Main", Line: 1, Column: 7},
		{Type: token.EOF, Line: 1, Column: 11},
	}
	p := New(toks)

	if !p.check(token.CLASS) {
		t.Fatal("peek deveria ver CLASS")
	}
	if got := p.next(); got.Type != token.CLASS {
		t.Fatalf("next deveria devolver CLASS, deu %s", got.Type)
	}
	if !p.check(token.TYPEID) {
		t.Fatal("após 1 next, peek deveria ver TYPEID")
	}
	if _, err := p.expect(token.TYPEID); err != nil {
		t.Fatalf("expect(TYPEID) não deveria falhar: %v", err)
	}
	if !p.atEnd() {
		t.Fatal("após consumir tudo, deveria estar no EOF")
	}
	if _, err := p.expect(token.CLASS); err == nil {
		t.Fatal("expect(CLASS) no EOF deveria falhar")
	}
}
