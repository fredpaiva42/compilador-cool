// Package codegen traduz a AST para um modulo WASM em formato de texto.
package codegen

import (
	"fmt"
	"strings"

	"cool/ast"
	"cool/runtime"
	"cool/semantic"
)

// Emitter acumula o texto do modulo em construcao.
type Emitter struct {
	b     strings.Builder
	sem   *semantic.Result
	level int // unidades de indentacao de dois espacos
}

// Generate devolve o modulo WAT completo do programa: preludio do runtime,
// funcoes dos metodos e _start.
func Generate(prog *ast.Program, sem *semantic.Result) (string, error) {
	e := &Emitter{sem: sem}
	e.b.WriteString(runtime.Prelude())
	for _, c := range prog.Classes {
		for _, f := range c.Features {
			m, ok := f.(*ast.Method)
			if !ok {
				continue // atributo vira campo do objeto, nao funcao
			}
			if err := e.method(c, m); err != nil {
				return "", err
			}
		}
	}
	e.start(prog)
	e.level = 0
	e.line(")")
	return e.b.String(), nil
}

// watType: Int ocupa i64 na pilha; Bool, String, objetos e SELF_TYPE sao i32
// (endereco ou 0/1).
func watType(name string) string {
	if name == "Int" {
		return "i64"
	}
	return "i32"
}

func (e *Emitter) method(c *ast.Class, m *ast.Method) error {
	decl := fmt.Sprintf("(func $%s_%s (param $self i32)", c.Name, m.Name)
	for _, f := range m.Formals {
		decl += fmt.Sprintf(" (param $%s %s)", f.Name, watType(f.Type))
	}
	decl += fmt.Sprintf(" (result %s)", watType(m.ReturnType))
	e.b.WriteString("\n")
	e.level = 1
	e.line("%s", decl)
	e.level = 2
	if err := e.expr(m.Body); err != nil {
		return fmt.Errorf("%s.%s: %w", c.Name, m.Name, err)
	}
	e.level = 1
	e.line(")")
	return nil
}

// expr emite as instrucoes que deixam o valor de x na pilha.
func (e *Emitter) expr(x ast.Expr) error {
	switch n := x.(type) {
	case *ast.IntConst:
		e.line("i64.const %s", n.Value)
	case *ast.BoolConst:
		if n.Value {
			e.line("i32.const 1")
		} else {
			e.line("i32.const 0")
		}
	case *ast.Neg:
		// o WASM nao tem negacao de inteiro: ~x e 0 - x
		e.line("i64.const 0")
		if err := e.expr(n.Expr); err != nil {
			return err
		}
		e.line("i64.sub")
	case *ast.Not:
		if err := e.expr(n.Expr); err != nil {
			return err
		}
		e.line("i32.eqz")
	case *ast.BinOp:
		return e.binOp(n)
	case *ast.Block:
		// o valor do bloco e o da ultima expressao; as anteriores caem na pilha
		for i, sub := range n.Exprs {
			if err := e.expr(sub); err != nil {
				return err
			}
			if i < len(n.Exprs)-1 {
				e.line("drop")
			}
		}
	default:
		return fmt.Errorf("expressao ainda nao suportada (%T)", x)
	}
	return nil
}

func (e *Emitter) binOp(n *ast.BinOp) error {
	if err := e.expr(n.Left); err != nil {
		return err
	}
	if err := e.expr(n.Right); err != nil {
		return err
	}
	switch n.Op {
	case "+":
		e.line("i64.add")
	case "-":
		e.line("i64.sub")
	case "*":
		e.line("i64.mul")
	case "/":
		e.line("i64.div_s") // inteira, truncando para zero
	case "<":
		e.line("i64.lt_s")
	case "<=":
		e.line("i64.le_s")
	case "=":
		return e.equality(n)
	default:
		return fmt.Errorf("operador ainda nao suportado (%q)", n.Op)
	}
	return nil
}

// equality escolhe a comparacao pelo tipo do operando esquerdo: Int ocupa i64,
// e Bool, String e objetos sao i32 (para objetos, comparacao de endereco).
// SELF_TYPE cai no i32, que e o certo: e um endereco de objeto.
func (e *Emitter) equality(n *ast.BinOp) error {
	if e.sem.Types[n.Left] == "Int" {
		e.line("i64.eq")
		return nil
	}
	e.line("i32.eq")
	return nil
}

// start monta a entrada do programa, que o manual define como (new Main).main().
// Duas simplificacoes desta etapa: o self provisorio e 0, porque alocacao de
// objetos ainda nao existe; e o resultado de main sai impresso com putint,
// instrumentacao que some quando o out_int do IO estiver pronto.
func (e *Emitter) start(prog *ast.Program) {
	e.b.WriteString("\n")
	e.level = 1
	e.line(`(func (export "_start")`)
	e.level = 2
	e.line("i32.const 0")
	e.line("call $Main_main")
	if mainReturnType(prog) == "Int" {
		e.line("call $putint")
	} else {
		e.line("drop")
	}
	e.line("i32.const 0")
	e.line("call $proc_exit")
	e.level = 1
	e.line(")")
}

func mainReturnType(prog *ast.Program) string {
	for _, c := range prog.Classes {
		if c.Name != "Main" {
			continue
		}
		for _, f := range c.Features {
			if m, ok := f.(*ast.Method); ok && m.Name == "main" {
				return m.ReturnType
			}
		}
	}
	return ""
}

func (e *Emitter) line(format string, args ...any) {
	e.b.WriteString(strings.Repeat("  ", e.level))
	fmt.Fprintf(&e.b, format, args...)
	e.b.WriteString("\n")
}
