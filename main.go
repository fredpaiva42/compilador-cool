package main

import (
	"fmt"
	"os"

	"cool/lexer"
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
	for {
		tok := l.NextToken()
		fmt.Printf("%d:%d\t%-10s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
		if tok.Type == token.EOF {
			break
		}
	}
}
