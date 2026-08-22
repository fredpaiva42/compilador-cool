package token

import "fmt"

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
