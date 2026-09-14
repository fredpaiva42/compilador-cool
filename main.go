package main

import (
	"fmt"
	"os"

	"cool/lexer"
	"cool/parser"
	"cool/token"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: go run . <arquivo.cl>")
		os.Exit(1)
	}

	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao ler arquivo: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(src))
	var toks []token.Token
	for {
		tok := l.NextToken()
		toks = append(toks, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	for _, tok := range toks {
		fmt.Printf("%d:%d\t%-10s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
	}

	for _, tok := range toks {
		if tok.Type == token.ILLEGAL {
			fmt.Fprintf(os.Stderr, "erro lexico em %d:%d: %s\n", tok.Line, tok.Column, tok.Literal)
			os.Exit(1)
		}
	}

	prog, err := parser.New(toks).ParseProgram()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro sintatico: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "OK: %d classe(s)\n", len(prog.Classes))
}
