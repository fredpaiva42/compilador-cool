package lexer

import (
	"cool/token"
	"strings"
)

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

	for {
		l.skipWhitespace()

		if l.pos >= len(l.input) {
			return token.Token{Type: token.EOF, Line: l.line, Column: l.column}
		}

		ch := l.input[l.pos]

		if ch == '-' && l.peekChar() == '-' {
			l.skipLineComment()
			continue
		}

		if ch == '(' && l.peekChar() == '*' {
			errLine, errCol := l.line, l.column

			if !l.skipBlockComment() {
				return token.Token{
					Type:    token.ILLEGAL,
					Literal: "comentário (* não fechado",
					Line:    errLine,
					Column:  errCol,
				}
			}
			continue
		}

		break
	}

	startLine, startCol := l.line, l.column
	startPos := l.pos

	ch := l.input[l.pos]
	next := l.peekChar()
	l.readChar()

	tok := token.Token{
		Type:    token.ILLEGAL,
		Literal: string(ch),
		Line:    startLine,
		Column:  startCol,
	}

	switch ch {
	case '(':
		tok.Type = token.LPAREN
	case ')':
		tok.Type = token.RPAREN
	case '{':
		tok.Type = token.LBRACE
	case '}':
		tok.Type = token.RBRACE
	case ':':
		tok.Type = token.COLON
	case ';':
		tok.Type = token.SEMI
	case '.':
		tok.Type = token.DOT
	case ',':
		tok.Type = token.COMMA
	case '@':
		tok.Type = token.AT
	case '+':
		tok.Type = token.PLUS
	case '-':
		tok.Type = token.MINUS
	case '*':
		tok.Type = token.MULT
	case '/':
		tok.Type = token.DIV
	case '~':
		tok.Type = token.NEG
	case '<':
		switch next {
		case '-':
			l.readChar()
			tok.Type = token.ASSIGN
			tok.Literal = "<-"
		case '=':
			l.readChar()
			tok.Type = token.LE
			tok.Literal = "<="
		default:
			tok.Type = token.LT
		}
	case '=':
		if next == '>' {
			l.readChar()
			tok.Type = token.DARROW
			tok.Literal = "=>"
		} else {
			tok.Type = token.EQ
		}
	case '"':
		content, errMsg := l.readString()
		if errMsg != "" {
			tok.Type = token.ILLEGAL
			tok.Literal = errMsg
			return tok
		}
		tok.Type = token.STR_CONST
		tok.Literal = content
		return tok
	default:
		switch {
		case isDigit(ch):
			l.readNumber()
			tok.Type = token.INT_CONST
			tok.Literal = l.input[startPos:l.pos]
		case isLetter(ch):
			l.readIdentifier()
			lit := l.input[startPos:l.pos]
			tok.Literal = lit

			if t, ok := token.LookupIdent(lit); ok {
				tok.Type = t // keyword ou booleano
			} else if 'A' <= ch && ch <= 'Z' {
				tok.Type = token.TYPEID
			} else {
				tok.Type = token.OBJECTID
			}
		}
	}

	return tok
}

func (l *Lexer) skipLineComment() {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.readChar()
	}
}

func (l *Lexer) skipBlockComment() bool {
	depth := 0

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		next := l.peekChar()

		if ch == '(' && next == '*' {
			depth++
			l.readChar()
			l.readChar()
			continue
		}

		if ch == '*' && next == ')' {
			depth--
			l.readChar()
			l.readChar()
			if depth == 0 {
				return true
			}
			continue
		}

		l.readChar()
	}

	return false

}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumber() {
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func (l *Lexer) readIdentifier() {
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.readChar()
	}
}

func (l *Lexer) readString() (content, errMsg string) {
	var sb strings.Builder

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		switch {
		case ch == '"':
			l.readChar()
			return sb.String(), ""

		case ch == '\\':
			l.readChar()
			if l.pos >= len(l.input) {
				return "", "string não fechada (barra solta no fim do arquivo)"
			}

			switch esc := l.input[l.pos]; esc {
			case 'b':
				sb.WriteByte('\b')
			case 't':
				sb.WriteByte('\t')
			case 'n':
				sb.WriteByte('\n')
			case 'f':
				sb.WriteByte('\f')
			default:
				sb.WriteByte(esc)
			}
			l.readChar()

		case ch == '\n':
			return "", "newline cru dentro de string"

		case ch == 0:
			return "", "caractere nulo (\\0) dentro de string"

		default:
			sb.WriteByte(ch)
			l.readChar()
		}

		if sb.Len() > 1024 {
			return "", "string exced o limite de 1024 caracteres"
		}
	}

	return "", "string não fechada (EOF antes das aspas finais)"
}
