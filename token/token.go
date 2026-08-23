package token

import (
	"fmt"
	"strings"
)

type Type int

type Token struct {
	Type    Type
	Literal string
	Line    int
	Column  int
}

const (
	// Especiais
	ILLEGAL Type = iota
	EOF

	// Literais e identificadores
	INT_CONST
	STR_CONST
	BOOL_CONST
	TYPEID   // Main, Int, SELF_TYPE
	OBJECTID // x, self, out_string

	// keywords
	CLASS
	ELSE
	FI
	IF
	IN
	INHERITS
	ISVOID
	LET
	LOOP
	POOL
	THEN
	WHILE
	CASE
	ESAC
	NEW
	OF
	NOT

	// Operadores e pontuação
	PLUS
	MINUS
	MULT
	DIV
	NEG
	LT
	LE
	EQ
	ASSIGN
	DARROW
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	COLON
	SEMI
	DOT
	COMMA
	AT
)

var typeNames = map[Type]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	INT_CONST:  "INT_CONST",
	STR_CONST:  "STR_CONST",
	BOOL_CONST: "BOOL_CONST",
	TYPEID:     "TYPEID",
	OBJECTID:   "OBJECTID",
	CLASS:      "CLASS",
	ELSE:       "ELSE",
	FI:         "FI",
	IF:         "IF",
	IN:         "IN",
	INHERITS:   "INHERITS",
	ISVOID:     "ISVOID",
	LET:        "LET",
	LOOP:       "LOOP",
	POOL:       "POOL",
	THEN:       "THEN",
	WHILE:      "WHILE",
	CASE:       "CASE",
	ESAC:       "ESAC",
	NEW:        "NEW",
	OF:         "OF",
	NOT:        "NOT",
	PLUS:       "PLUS",
	MINUS:      "MINUS",
	MULT:       "MULT",
	DIV:        "DIV",
	NEG:        "NEG",
	LT:         "LT",
	LE:         "LE",
	EQ:         "EQ",
	ASSIGN:     "ASSIGN",
	DARROW:     "DARROW",
	LPAREN:     "LPAREN",
	RPAREN:     "RPAREN",
	LBRACE:     "LBRACE",
	RBRACE:     "RBRACE",
	COLON:      "COLON",
	SEMI:       "SEMI",
	DOT:        "DOT",
	COMMA:      "COMMA",
	AT:         "AT",
}

func (t Type) String() string {
	if name, ok := typeNames[t]; ok {
		return name
	}

	return fmt.Sprintf("Type(%d)", int(t))
}

var keywords = map[string]Type{
	"class":    CLASS,
	"else":     ELSE,
	"fi":       FI,
	"if":       IF,
	"in":       IN,
	"inherits": INHERITS,
	"isvoid":   ISVOID,
	"let":      LET,
	"loop":     LOOP,
	"pool":     POOL,
	"then":     THEN,
	"while":    WHILE,
	"case":     CASE,
	"esac":     ESAC,
	"new":      NEW,
	"of":       OF,
	"not":      NOT,
	"true":     BOOL_CONST,
	"false":    BOOL_CONST,
}

func LookupIdent(ident string) (Type, bool) {
	t, ok := keywords[strings.ToLower(ident)]
	if !ok {
		return t, false
	}

	if t == BOOL_CONST {
		if ident[0] != 't' && ident[0] != 'f' {
			return t, false
		}
	}

	return t, true
}
