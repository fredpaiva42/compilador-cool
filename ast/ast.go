package ast

type Program struct {
	Classes []*Class
}

type Class struct {
	Name     string
	Parent   string
	Features []Feature
	Line     int
	Column   int
}

type Feature interface {
	featureNode()
}

type Formal struct {
	Name   string
	Type   string
	Line   int
	Column int
}

type Attribute struct {
	Name    string
	Type    string
	Init    Expr
	HasInit bool
	Line    int
	Column  int
}

type Method struct {
	Name       string
	Formals    []*Formal
	ReturnType string
	Body       Expr
	Line       int
	Column     int
}

func (*Attribute) featureNode() {}

func (*Method) featureNode() {}

type Expr interface {
	exprNode()
}
