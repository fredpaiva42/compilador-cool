package codegen

import (
	"fmt"
	"strings"

	"cool/ast"
	"cool/runtime"
)

// Emitter acumula o texto do modulo em construcao.
type Emitter struct {
	b strings.Builder
}

// Gerar devolve o modulo WAT completo do programa: preludio do runtime,
// funcoes dos metodos e _start.
func Gerar(prog *ast.Program) (string, error) {
	e := &Emitter{}
	e.b.WriteString(runtime.Preludio())
	for _, c := range prog.Classes {
		for _, f := range c.Features {
			m, ok := f.(*ast.Method)
			if !ok {
				continue // atributo vira campo do objeto, nao funcao
			}
			if err := e.metodo(c, m); err != nil {
				return "", err
			}
		}
	}
	e.start(prog)
	e.b.WriteString(")\n")
	return e.b.String(), nil
}

// tipoWAT: Int ocupa i64 na pilha; Bool, String, objetos e SELF_TYPE sao i32
// (endereco ou 0/1).
func tipoWAT(nome string) string {
	if nome == "Int" {
		return "i64"
	}
	return "i32"
}

func (e *Emitter) metodo(c *ast.Class, m *ast.Method) error {
	fmt.Fprintf(&e.b, "\n  (func $%s_%s (param $self i32)", c.Name, m.Name)
	for _, f := range m.Formals {
		fmt.Fprintf(&e.b, " (param $%s %s)", f.Name, tipoWAT(f.Type))
	}
	fmt.Fprintf(&e.b, " (result %s)\n", tipoWAT(m.ReturnType))
	if err := e.expr(m.Body); err != nil {
		return fmt.Errorf("%s.%s: %w", c.Name, m.Name, err)
	}
	e.b.WriteString("  )\n")
	return nil
}

func (e *Emitter) expr(x ast.Expr) error {
	switch n := x.(type) {
	case *ast.IntConst:
		e.linha("(i64.const %s)", n.Value)
		return nil
	case *ast.BoolConst:
		if n.Value {
			e.linha("(i32.const 1)")
		} else {
			e.linha("(i32.const 0)")
		}
		return nil
	default:
		return fmt.Errorf("expressao ainda nao suportada (%T)", x)
	}
}

// start monta a entrada do programa, que o manual define como (new Main).main().
// Duas simplificacoes desta etapa: o self provisorio e 0, porque alocacao de
// objetos ainda nao existe; e o resultado de main sai impresso com putint,
// instrumentacao que some quando o out_int do IO estiver pronto.
func (e *Emitter) start(prog *ast.Program) {
	e.b.WriteString("\n  (func (export \"_start\")\n")
	if tipoRetornoMain(prog) == "Int" {
		e.linha("(call $putint (call $Main_main (i32.const 0)))")
	} else {
		e.linha("(drop (call $Main_main (i32.const 0)))")
	}
	e.linha("(call $proc_exit (i32.const 0))")
	e.b.WriteString("  )\n")
}

func tipoRetornoMain(prog *ast.Program) string {
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

func (e *Emitter) linha(format string, args ...any) {
	e.b.WriteString("    ")
	fmt.Fprintf(&e.b, format, args...)
	e.b.WriteString("\n")
}
