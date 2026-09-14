package main

import (
	"fmt"
	"os"

	"cool/ast"
	"cool/lexer"
	"cool/parser"
	"cool/token"
)

func main() {
	withAST := false
	fileArg := ""
	if len(os.Args) >= 3 && os.Args[1] == "--ast" {
		withAST = true
		fileArg = os.Args[2]
	} else if len(os.Args) >= 2 {
		fileArg = os.Args[1]
	} else {
		fmt.Fprintln(os.Stderr, "uso: go run . [--ast] <arquivo.cl>")
		os.Exit(1)
	}

	src, err := os.ReadFile(fileArg)
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

	if withAST {
		fmt.Println(ast.DumpProgram(prog))
	}

	fmt.Fprintf(os.Stderr, "OK: %d classe(s)\n", len(prog.Classes))
}
