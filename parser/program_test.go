package parser

import (
	"cool/lexer"
	"cool/token"
	"testing"
)

func tokenize(src string) []token.Token {
	l := lexer.New(src)
	var toks []token.Token
	for {
		tok := l.NextToken()
		toks = append(toks, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	return toks
}

func TestEmptyClass(t *testing.T) {
	prog, err := New(tokenize("class Main { };")).ParseProgram()
	if err != nil {
		t.Fatalf("não deveria falhar: %v", err)
	}

	if len(prog.Classes) != 1 {
		t.Fatalf("esperava 1 classe, deu %d", len(prog.Classes))
	}

	c := prog.Classes[0]
	if c.Name != "Main" || c.Parent != "" || len(c.Features) != 0 {
		t.Fatalf("classe errada: %+v", c)
	}
}

func TestEmptyClassWithInherits(t *testing.T) {
	prog, err := New(tokenize("class Dog inherits Animal { };")).ParseProgram()
	if err != nil {
		t.Fatalf("não deveria falhar: %v", err)
	}
	c := prog.Classes[0]
	if c.Name != "Dog" || c.Parent != "Animal" {
		t.Fatalf("herança errada: %+v", c)
	}
}
