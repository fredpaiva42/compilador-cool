package lexer

import "cool/token"

type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

func New(input string) *Lexer {
	return &Lexer{input: input, line: 1, column: 1}
}

func (l *Lexer) readChar() {
	if l.input[l.pos] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
}

func (l *Lexer) peekChar() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		switch l.input[l.pos] {
		case ' ', '\t', '\n', '\r', '\f', '\v':
			l.readChar()
		default:
			return
		}
	}
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	startLine, startCol := l.line, l.column

	if l.pos >= len(l.input) {
		return token.Token{Type: token.EOF, Line: startLine, Column: startCol}
	}

	ch := l.input[l.pos]
	l.readChar()
	return token.Token{Type: token.ILLEGAL, Literal: string(ch),
		Line: startLine, Column: startCol}
}
