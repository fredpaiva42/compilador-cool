package semantic

import (
	"testing"

	"cool/ast"
	"cool/lexer"
	"cool/parser"
	"cool/token"
)

// parseSrc roda lexer → parser e devolve a AST.
func parseSrc(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	var toks []token.Token
	for {
		tok := l.NextToken()
		toks = append(toks, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	prog, err := parser.New(toks).ParseProgram()
	if err != nil {
		t.Fatalf("parse falhou: %v", err)
	}
	return prog
}

func TestTypesRecorded(t *testing.T) {
	prog := parseSrc(t, `class Main {
  main() : Object { { 1 + 2; true; "oi"; let n : Int <- 3 in n; self; } };
};`)
	res, err := Analyze(prog)
	if err != nil {
		t.Fatalf("analise falhou: %v", err)
	}

	block := prog.Classes[0].Features[0].(*ast.Method).Body.(*ast.Block)
	cases := []struct {
		name string
		x    ast.Expr
		want string
	}{
		{"sum", block.Exprs[0], "Int"},
		{"bool", block.Exprs[1], "Bool"},
		{"string", block.Exprs[2], "String"},
		{"let", block.Exprs[3], "Int"},
		{"self", block.Exprs[4], "SELF_TYPE"},
	}
	for _, c := range cases {
		got, ok := res.Types[c.x]
		if !ok {
			t.Errorf("%s: tipo nao registrado", c.name)
			continue
		}
		if got != c.want {
			t.Errorf("%s: tipo %q, quer %q", c.name, got, c.want)
		}
	}

	// a recursão também registra as subexpressões
	sum := block.Exprs[0].(*ast.BinOp)
	for _, side := range []ast.Expr{sum.Left, sum.Right} {
		if res.Types[side] != "Int" {
			t.Errorf("operando da soma com tipo %q, quer Int", res.Types[side])
		}
	}
}

func TestAnalyzeReturnsTable(t *testing.T) {
	res, err := Analyze(parseSrc(t, `class Main inherits IO { main() : Object { out_int(1) }; };`))
	if err != nil {
		t.Fatalf("analise falhou: %v", err)
	}
	sig, ok := res.Tab.Methods["Main"]["main"]
	if !ok {
		t.Fatal("Main.main ausente da tabela")
	}
	if sig.Owner != "Main" || sig.Return != "Object" {
		t.Errorf("assinatura de Main.main: owner %q, retorno %q", sig.Owner, sig.Return)
	}
	if _, ok := res.Tab.Attrs["Main"]; !ok {
		t.Error("tabela sem entrada de atributos para Main")
	}
}
