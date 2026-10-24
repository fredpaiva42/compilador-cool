package main

import (
	"fmt"
	"os"
	"strings"

	"cool/ast"
	"cool/codegen"
	"cool/lexer"
	"cool/parser"
	"cool/semantic"
	"cool/token"
)

func main() {
	args := os.Args[1:]
	flag := ""
	if len(args) > 0 && strings.HasPrefix(args[0], "--") {
		flag = args[0]
		args = args[1:]
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "uso: go run . [--ast|--wat] <arquivo.cl>")
		os.Exit(1)
	}
	fileArg := args[0]

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

	if flag == "" {
		for _, tok := range toks {
			fmt.Printf("%d:%d\t%-10s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
		}
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

	if flag == "--ast" {
		fmt.Println(ast.DumpProgram(prog))
	}

	if err := semantic.CheckProgram(prog); err != nil {
		fmt.Fprintf(os.Stderr, "erro semantico: %v\n", err)
		os.Exit(1)
	}

	if flag == "--wat" {
		wat, err := codegen.Generate(prog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro na geracao: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(wat)
		return
	}

	fmt.Fprintf(os.Stderr, "OK: %d classe(s)\n", len(prog.Classes))
}
