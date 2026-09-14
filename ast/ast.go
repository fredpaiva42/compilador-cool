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

type IntConst struct {
	Value  string
	Line   int
	Column int
}

func (*IntConst) exprNode() {}

type StringConst struct {
	Value  string
	Line   int
	Column int
}

func (*StringConst) exprNode() {}

type BoolConst struct {
	Value  bool
	Line   int
	Column int
}

func (*BoolConst) exprNode() {}

type Object struct {
	Name   string
	Line   int
	Column int
}

func (*Object) exprNode() {}

type New struct {
	Type   string
	Line   int
	Column int
}

func (*New) exprNode() {}

type IsVoid struct {
	Expr   Expr
	Line   int
	Column int
}

func (*IsVoid) exprNode() {}

type Dispatch struct {
	Receiver Expr
	Method   string
	Args     []Expr
	Line     int
	Column   int
}

func (*Dispatch) exprNode() {}

type StaticDispatch struct {
	Receiver   Expr
	StaticType string
	Method     string
	Args       []Expr
	Line       int
	Column     int
}

func (*StaticDispatch) exprNode() {}

type If struct {
	Cond   Expr
	Then   Expr
	Else   Expr
	Line   int
	Column int
}

func (*If) exprNode() {}

type While struct {
	Cond   Expr
	Body   Expr
	Line   int
	Column int
}

func (*While) exprNode() {}

type Block struct {
	Exprs  []Expr
	Line   int
	Column int
}

func (*Block) exprNode() {}

type LetBinding struct {
	Name    string
	Type    string
	Init    Expr
	HasInit bool
	Line    int
	Column  int
}

type Let struct {
	Bindings []*LetBinding
	Body     Expr
	Line     int
	Column   int
}

func (*Let) exprNode() {}

type CaseBranch struct {
	Name   string
	Type   string
	Body   Expr
	Line   int
	Column int
}

type Case struct {
	Subject  Expr
	Branches []*CaseBranch
	Line     int
	Column   int
}

func (*Case) exprNode() {}

type Assign struct {
	Target string
	Value  Expr
	Line   int
	Column int
}

func (*Assign) exprNode() {}

type BinOp struct {
	Op     string
	Left   Expr
	Right  Expr
	Line   int
	Column int
}

func (*BinOp) exprNode() {}

type Neg struct {
	Expr   Expr
	Line   int
	Column int
}

func (*Neg) exprNode() {}

type Not struct {
	Expr   Expr
	Line   int
	Column int
}

func (*Not) exprNode() {}
