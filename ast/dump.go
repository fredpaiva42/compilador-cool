package ast

import (
	"fmt"
	"strings"
)

func DumpProgram(p *Program) string {
	var sb strings.Builder
	sb.WriteString("(program")
	for _, c := range p.Classes {
		sb.WriteString(" " + DumpClass(c))
	}
	sb.WriteString(")")
	return sb.String()
}

func DumpClass(c *Class) string {
	parent := c.Parent
	if parent == "" {
		parent = "Object"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "(class %s %s", c.Name, parent)
	for _, f := range c.Features {
		sb.WriteString(" " + dumpFeature(f))
	}
	sb.WriteString(")")
	return sb.String()
}

func dumpFeature(f Feature) string {
	switch n := f.(type) {
	case *Attribute:
		if n.HasInit {
			return fmt.Sprintf("(attr %s %s %s)", n.Name, n.Type, DumpExpr(n.Init))
		}
		return fmt.Sprintf("(attr %s %s)", n.Name, n.Type)
	case *Method:
		var sb strings.Builder
		fmt.Fprintf(&sb, "(method %s (", n.Name)
		for i, fm := range n.Formals {
			if i > 0 {
				sb.WriteString(" ")
			}
			fmt.Fprintf(&sb, "(%s %s)", fm.Name, fm.Type)
		}
		fmt.Fprintf(&sb, ") %s %s)", n.ReturnType, DumpExpr(n.Body))
		return sb.String()
	default:
		return "(feature?)"
	}
}

func dumpList(es []Expr) string {
	parts := make([]string, len(es))
	for i, e := range es {
		parts[i] = DumpExpr(e)
	}
	return strings.Join(parts, " ")
}

func DumpExpr(e Expr) string {
	switch n := e.(type) {
	case *IntConst:
		return n.Value
	case *StringConst:
		return fmt.Sprintf("%q", n.Value)
	case *BoolConst:
		if n.Value {
			return "true"
		}
		return "false"
	case *Object:
		return n.Name
	case *New:
		return "(new " + n.Type + ")"
	case *IsVoid:
		return "(isvoid " + DumpExpr(n.Expr) + ")"
	case *Dispatch:
		recv := "self"
		if n.Receiver != nil {
			recv = DumpExpr(n.Receiver)
		}
		if len(n.Args) == 0 {
			return fmt.Sprintf("(call %s %s)", recv, n.Method)
		}
		return fmt.Sprintf("(call %s %s %s)", recv, n.Method, dumpList(n.Args))
	case *StaticDispatch:
		if len(n.Args) == 0 {
			return fmt.Sprintf("(static-call %s %s %s)", DumpExpr(n.Receiver), n.StaticType, n.Method)
		}
		return fmt.Sprintf("(static-call %s %s %s %s)", DumpExpr(n.Receiver), n.StaticType, n.Method, dumpList(n.Args))
	case *Assign:
		return fmt.Sprintf("(assign %s %s)", n.Target, DumpExpr(n.Value))
	case *BinOp:
		return fmt.Sprintf("(%s %s %s)", n.Op, DumpExpr(n.Left), DumpExpr(n.Right))
	case *Neg:
		return "(~ " + DumpExpr(n.Expr) + ")"
	case *Not:
		return "(not " + DumpExpr(n.Expr) + ")"
	case *If:
		return fmt.Sprintf("(if %s %s %s)", DumpExpr(n.Cond), DumpExpr(n.Then), DumpExpr(n.Else))
	case *While:
		return fmt.Sprintf("(while %s %s)", DumpExpr(n.Cond), DumpExpr(n.Body))
	case *Block:
		return "(block " + dumpList(n.Exprs) + ")"
	case *Let:
		var sb strings.Builder
		sb.WriteString("(let (")
		for i, b := range n.Bindings {
			if i > 0 {
				sb.WriteString(" ")
			}
			if b.HasInit {
				fmt.Fprintf(&sb, "(%s %s %s)", b.Name, b.Type, DumpExpr(b.Init))
			} else {
				fmt.Fprintf(&sb, "(%s %s)", b.Name, b.Type)
			}
		}
		fmt.Fprintf(&sb, ") %s)", DumpExpr(n.Body))
		return sb.String()
	case *Case:
		var sb strings.Builder
		sb.WriteString("(case " + DumpExpr(n.Subject))
		for _, b := range n.Branches {
			fmt.Fprintf(&sb, " (%s %s %s)", b.Name, b.Type, DumpExpr(b.Body))
		}
		sb.WriteString(")")
		return sb.String()
	default:
		return "(?)"
	}
}
