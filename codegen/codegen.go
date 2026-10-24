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
	b      strings.Builder
	sem    *semantic.Result
	level  int                        // unidades de indentacao de dois espacos
	locals []string                   // declaracoes (local $l0 i64), uma por let
	bound  map[*ast.LetBinding]string // cada binding ganha um local proprio
	vars   map[string]string          // nome visivel -> local ou parametro
	class  string                     // classe do metodo em emissao
	labels int                        // contador de rotulos para o while
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
	e.locals = nil
	e.bound = map[*ast.LetBinding]string{}
	e.vars = map[string]string{}
	e.class = c.Name
	e.labels = 0
	for _, f := range m.Formals {
		e.vars[f.Name] = "$" + f.Name
	}
	e.allocLocals(m.Body)
	decl := fmt.Sprintf("(func $%s_%s (param $self i32)", c.Name, m.Name)
	for _, f := range m.Formals {
		decl += fmt.Sprintf(" (param $%s %s)", f.Name, watType(f.Type))
	}
	decl += fmt.Sprintf(" (result %s)", watType(m.ReturnType))
	e.b.WriteString("\n")
	e.level = 1
	e.line("%s", decl)
	e.level = 2
	for _, d := range e.locals {
		e.line("%s", d)
	}
	if err := e.expr(m.Body); err != nil {
		return fmt.Errorf("%s.%s: %w", c.Name, m.Name, err)
	}
	e.level = 1
	e.line(")")
	return nil
}

// allocLocals percorre o corpo antes da emissao e reserva um local WASM por
// binding de let. O WAT exige os locals declarados no inicio da funcao, entao
// nao da para criar sob demanda durante a emissao. Cada binding ganha um nome
// proprio ($l0, $l1, ...) para o sombreamento nao colidir.
func (e *Emitter) allocLocals(x ast.Expr) {
	switch n := x.(type) {
	case *ast.Let:
		for _, b := range n.Bindings {
			name := fmt.Sprintf("$l%d", len(e.locals))
			e.bound[b] = name
			e.locals = append(e.locals, fmt.Sprintf("(local %s %s)", name, watType(b.Type)))
			if b.HasInit {
				e.allocLocals(b.Init)
			}
		}
		e.allocLocals(n.Body)
	case *ast.Block:
		for _, sub := range n.Exprs {
			e.allocLocals(sub)
		}
	case *ast.If:
		e.allocLocals(n.Cond)
		e.allocLocals(n.Then)
		e.allocLocals(n.Else)
	case *ast.While:
		e.allocLocals(n.Cond)
		e.allocLocals(n.Body)
	case *ast.BinOp:
		e.allocLocals(n.Left)
		e.allocLocals(n.Right)
	case *ast.Neg:
		e.allocLocals(n.Expr)
	case *ast.Not:
		e.allocLocals(n.Expr)
	case *ast.IsVoid:
		e.allocLocals(n.Expr)
	case *ast.Assign:
		e.allocLocals(n.Value)
	}
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
	case *ast.Object:
		return e.object(n)
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
	case *ast.IsVoid:
		return e.isVoid(n)
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
	case *ast.If:
		return e.branch(n)
	case *ast.While:
		return e.loop(n)
	case *ast.Let:
		return e.let(n)
	case *ast.Assign:
		return e.assign(n)
	default:
		return fmt.Errorf("expressao ainda nao suportada (%T)", x)
	}
	return nil
}

// object resolve um nome: self e o parametro implicito, o resto e local de
// let ou formal. Atributo nu ainda nao sai: vira campo do objeto no passo 4.
func (e *Emitter) object(n *ast.Object) error {
	if n.Name == "self" {
		e.line("local.get $self")
		return nil
	}
	if loc, ok := e.vars[n.Name]; ok {
		e.line("local.get %s", loc)
		return nil
	}
	return fmt.Errorf("atributo %s ainda nao suportado (passo 4)", n.Name)
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

// branch emite o if com tipo de resultado: os dois ramos precisam deixar o
// mesmo tipo WASM na pilha. Ramos de tipos diferentes (um Int, outro objeto)
// so funcionam com boxing, que chega no passo 4.
func (e *Emitter) branch(n *ast.If) error {
	rt := watType(e.sem.Types[n])
	if watType(e.sem.Types[n.Then]) != rt || watType(e.sem.Types[n.Else]) != rt {
		return fmt.Errorf("if com ramos de tipos diferentes ainda nao suportado (passo 4)")
	}
	if err := e.expr(n.Cond); err != nil {
		return err
	}
	e.line("(if (result %s)", rt)
	e.line("(then")
	e.level++
	if err := e.expr(n.Then); err != nil {
		return err
	}
	e.level--
	e.line(")")
	e.line("(else")
	e.level++
	if err := e.expr(n.Else); err != nil {
		return err
	}
	e.level--
	e.line(")")
	e.line(")")
	return nil
}

// loop emite o while como block+loop: testa a condicao, sai quando e falsa,
// avalia o corpo e descarta o valor, e repete. O while tipa Object no COOL,
// entao o valor final e 0 (void).
func (e *Emitter) loop(n *ast.While) error {
	w := e.labels
	e.labels++
	e.line("(block $w%d", w)
	e.level++
	e.line("(loop $l%d", w)
	e.level++
	if err := e.expr(n.Cond); err != nil {
		return err
	}
	e.line("i32.eqz")
	e.line("br_if $w%d", w)
	if err := e.expr(n.Body); err != nil {
		return err
	}
	e.line("drop")
	e.line("br $l%d", w)
	e.level--
	e.line(")")
	e.level--
	e.line(")")
	e.line("i32.const 0")
	return nil
}

// let avalia cada init na ordem e guarda no local proprio do binding; os
// bindings seguintes enxergam os anteriores, e o corpo enxerga todos. Nomes
// repetidos ganham locais distintos, e o anterior volta a valer depois.
func (e *Emitter) let(n *ast.Let) error {
	type saved struct {
		name  string
		local string
		ok    bool
	}
	var stack []saved
	for _, b := range n.Bindings {
		loc := e.bound[b]
		if b.HasInit {
			if watType(e.sem.Types[b.Init]) != watType(b.Type) {
				return fmt.Errorf("let %s com init de outro tipo ainda nao suportado (passo 4)", b.Name)
			}
			if err := e.expr(b.Init); err != nil {
				return err
			}
		} else {
			e.defaultValue(b.Type)
		}
		e.line("local.set %s", loc)
		prev, ok := e.vars[b.Name]
		stack = append(stack, saved{b.Name, prev, ok})
		e.vars[b.Name] = loc
	}
	if err := e.expr(n.Body); err != nil {
		return err
	}
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].ok {
			e.vars[stack[i].name] = stack[i].local
		} else {
			delete(e.vars, stack[i].name)
		}
	}
	return nil
}

// defaultValue empilha o valor padrao do COOL: 0 para Int, falso para Bool,
// void (endereco 0) para o resto.
func (e *Emitter) defaultValue(typ string) {
	if typ == "Int" {
		e.line("i64.const 0")
	} else {
		e.line("i32.const 0")
	}
}

// assign avalia o valor e guarda no local, mantendo uma copia na pilha: pela
// regra [ASSIGN] a atribuicao vale o proprio valor.
func (e *Emitter) assign(n *ast.Assign) error {
	if err := e.expr(n.Value); err != nil {
		return err
	}
	loc, ok := e.vars[n.Target]
	if !ok {
		return fmt.Errorf("atribuicao a %s ainda nao suportada (passo 4)", n.Target)
	}
	e.line("local.tee %s", loc)
	return nil
}

// isVoid: Int e Bool nunca sao void, entao o valor sai da pilha e entra
// falso; endereco (String, objetos, SELF_TYPE) testa contra zero.
func (e *Emitter) isVoid(n *ast.IsVoid) error {
	t := e.sem.Types[n.Expr]
	if err := e.expr(n.Expr); err != nil {
		return err
	}
	if t == "Int" || t == "Bool" {
		e.line("drop")
		e.line("i32.const 0")
		return nil
	}
	e.line("i32.eqz")
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
