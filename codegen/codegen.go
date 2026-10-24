// Package codegen traduz a AST para um modulo WASM em formato de texto.
//
// Modelo de objetos (passo 4): todo valor COOL e um endereco i32, e 0 e void.
// Cada objeto tem cabecalho de 8 bytes (tag em +0, tamanho em +4): Int guarda
// o i64 em +8, Bool o i32 em +8, objeto de usuario um campo de 8 bytes por
// atributo a partir de +8. O boxing deixa if, let e assign uniformes: todos
// empilham e guardam enderecos, sem checagem de tipo na emissao.
package codegen

import (
	"fmt"
	"sort"
	"strings"

	"cool/ast"
	"cool/runtime"
	"cool/semantic"
)

// Tags das classes basicas; as do programa vao de 6 em ordem alfabetica.
const (
	tagObject = 1
	tagIO     = 2
	tagInt    = 3
	tagString = 4
	tagBool   = 5
)

// Emitter acumula o texto do modulo em construcao.
type Emitter struct {
	b       strings.Builder
	sem     *semantic.Result
	level   int                        // unidades de indentacao de dois espacos
	locals  []string                   // declaracoes (local $l0 i32), uma por let
	bound   map[*ast.LetBinding]string // cada binding ganha um local proprio
	vars    map[string]string          // nome visivel -> local ou parametro
	class   string                     // classe do metodo em emissao
	labels  int                        // contador de rotulos para o while
	tags    map[string]int             // classe -> tag
	scratch int                        // maior numero de args de despacho ($a0...)
}

// Generate devolve o modulo WAT completo do programa: preludio do runtime,
// construtores, funcoes dos metodos e _start.
func Generate(prog *ast.Program, sem *semantic.Result) (string, error) {
	e := &Emitter{sem: sem, tags: classTags(prog)}
	e.b.WriteString(runtime.Prelude())
	e.objectBuiltins()
	for _, c := range prog.Classes {
		if err := e.constructor(c); err != nil {
			return "", err
		}
	}
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

// objectBuiltins sintetiza os metodos de Object implementados no runtime:
// abort sai com codigo 1 e copy duplica o objeto. type_name fica para o
// passo 7 e os de IO/String para o 6. Nao colide com override: o metodo do
// usuario sai como $Classe_metodo, este como $Object_metodo.
func (e *Emitter) objectBuiltins() {
	e.b.WriteString("\n")
	e.level = 1
	e.line("(func $Object_abort (param $self i32) (result i32)")
	e.level = 2
	e.line("call $abort")
	e.line("unreachable")
	e.level = 1
	e.line(")")
	e.line("(func $Object_copy (param $self i32) (result i32)")
	e.level = 2
	e.line("local.get $self")
	e.line("call $copy_obj")
	e.level = 1
	e.line(")")
}

// classTags numera as classes do programa de 6 em diante, em ordem
// alfabetica: a ordem de declaracao nao serve porque mapa do Go embaralha.
func classTags(prog *ast.Program) map[string]int {
	tags := map[string]int{
		"Object": tagObject, "IO": tagIO, "Int": tagInt,
		"String": tagString, "Bool": tagBool,
	}
	var names []string
	for _, c := range prog.Classes {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	for i, n := range names {
		tags[n] = 6 + i
	}
	return tags
}

// attrField e um atributo com o tipo declarado.
type attrField struct {
	name string
	typ  string
}

// attrLayout lista os campos na ordem pai-antes-do-filho, e dentro da classe
// na ordem de declaracao: e a ordem em que o construtor inicializa.
func attrLayout(tab *semantic.Table, className string) []attrField {
	var chain []string
	for c := className; ; {
		chain = append(chain, c)
		info := tab.Classes[c]
		p := info.Parent
		if p == "" {
			if c == "Object" {
				break
			}
			p = "Object" // heranca implicita
		}
		c = p
	}
	var out []attrField
	for i := len(chain) - 1; i >= 0; i-- {
		info := tab.Classes[chain[i]]
		if info.Node == nil {
			continue // basica: sem atributos
		}
		for _, f := range info.Node.Features {
			if a, ok := f.(*ast.Attribute); ok {
				out = append(out, attrField{a.Name, a.Type})
			}
		}
	}
	return out
}

// attrOffset devolve o deslocamento do campo no objeto, ou -1 se nao existe.
func (e *Emitter) attrOffset(attr string) int {
	for i, f := range attrLayout(e.sem.Tab, e.class) {
		if f.name == attr {
			return 8 + 8*i
		}
	}
	return -1
}

// subclasses lista as classes que conformam com class, em ordem de tag:
// serve ao new SELF_TYPE e, no passo 5, ao despacho.
func (e *Emitter) subclasses(class string) []string {
	var out []string
	for name := range e.sem.Tab.Classes {
		if e.sem.Tab.Conforms(name, class) {
			out = append(out, name)
		}
	}
	sort.Slice(out, func(i, j int) bool { return e.tags[out[i]] < e.tags[out[j]] })
	return out
}

func (e *Emitter) method(c *ast.Class, m *ast.Method) error {
	e.beginScope(c.Name)
	for _, f := range m.Formals {
		e.vars[f.Name] = "$" + f.Name
	}
	e.allocLocals(m.Body)
	e.b.WriteString("\n")
	e.level = 1
	decl := fmt.Sprintf("(func $%s_%s (param $self i32)", c.Name, m.Name)
	for _, f := range m.Formals {
		decl += fmt.Sprintf(" (param $%s i32)", f.Name)
	}
	decl += " (result i32)"
	e.line("%s", decl)
	e.level = 2
	e.line("(local $tmp0 i32)")
	e.scratchDecls()
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

// scratchDecls declara um temporario por posicao de argumento ($a0...): o
// despacho dinamico avalia os args uma vez e guarda antes da ramificacao.
func (e *Emitter) scratchDecls() {
	for i := 0; i < e.scratch; i++ {
		e.line("(local $a%d i32)", i)
	}
}

// beginScope zera o estado por funcao: labels do WAT so valem dentro dela.
func (e *Emitter) beginScope(class string) {
	e.locals = nil
	e.bound = map[*ast.LetBinding]string{}
	e.vars = map[string]string{}
	e.class = class
	e.labels = 0
	e.scratch = 0
}

// constructor emite $C_new: aloca com zeros, poe os padroes e avalia os inits
// na ordem do layout, com self apontando para o objeto em construcao.
func (e *Emitter) constructor(c *ast.Class) error {
	fields := attrLayout(e.sem.Tab, c.Name)
	size := 8 + 8*len(fields)
	e.beginScope(c.Name)
	inits := map[string]*ast.Attribute{}
	for _, f := range fields {
		if a := findAttr(e.sem.Tab, c.Name, f.name); a != nil && a.HasInit {
			inits[f.name] = a
			e.allocLocals(a.Init)
		}
	}
	e.b.WriteString("\n")
	e.level = 1
	e.line("(func $%s_new (result i32)", c.Name)
	e.level = 2
	e.line("(local $self i32)")
	e.line("(local $tmp0 i32)")
	e.scratchDecls()
	for _, d := range e.locals {
		e.line("%s", d)
	}
	e.line("i32.const %d", e.tags[c.Name])
	e.line("i32.const %d", size)
	e.line("call $obj_new")
	e.line("local.set $self")
	for i, f := range fields {
		e.line("local.get $self")
		e.defaultValue(f.typ)
		e.line("i32.store offset=%d", 8+8*i)
	}
	for i, f := range fields {
		if a, ok := inits[f.name]; ok {
			e.line("local.get $self")
			if err := e.expr(a.Init); err != nil {
				return fmt.Errorf("%s.%s: %w", c.Name, f.name, err)
			}
			e.line("i32.store offset=%d", 8+8*i)
		}
	}
	e.line("local.get $self")
	e.level = 1
	e.line(")")
	return nil
}

// findAttr acha o atributo na classe que o declara, subindo a hierarquia.
func findAttr(tab *semantic.Table, className, attr string) *ast.Attribute {
	for c := className; ; {
		info := tab.Classes[c]
		if info.Node != nil {
			for _, f := range info.Node.Features {
				if a, ok := f.(*ast.Attribute); ok && a.Name == attr {
					return a
				}
			}
		}
		p := info.Parent
		if p == "" {
			if c == "Object" {
				return nil
			}
			p = "Object"
		}
		c = p
	}
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
			e.locals = append(e.locals, fmt.Sprintf("(local %s i32)", name))
			if b.HasInit {
				e.allocLocals(b.Init)
			}
		}
		e.allocLocals(n.Body)
	case *ast.New:
		// new nao tem subexpressoes
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
	case *ast.Dispatch:
		if n.Receiver != nil {
			e.allocLocals(n.Receiver)
		}
		for _, a := range n.Args {
			e.allocLocals(a)
		}
		if len(n.Args) > e.scratch {
			e.scratch = len(n.Args)
		}
	case *ast.StaticDispatch:
		e.allocLocals(n.Receiver)
		for _, a := range n.Args {
			e.allocLocals(a)
		}
		if len(n.Args) > e.scratch {
			e.scratch = len(n.Args)
		}
	case *ast.Case:
		e.allocLocals(n.Subject)
		for _, b := range n.Branches {
			e.allocLocals(b.Body)
		}
	}
}

// expr emite as instrucoes que deixam o valor de x na pilha: sempre um
// endereco i32 (0 = void).
func (e *Emitter) expr(x ast.Expr) error {
	switch n := x.(type) {
	case *ast.IntConst:
		e.line("i64.const %s", n.Value)
		e.line("call $new_int")
	case *ast.BoolConst:
		if n.Value {
			e.line("i32.const 1")
		} else {
			e.line("i32.const 0")
		}
		e.line("call $new_bool")
	case *ast.Object:
		return e.object(n)
	case *ast.New:
		return e.newObj(n)
	case *ast.Neg:
		// o WASM nao tem negacao de inteiro: ~x e 0 - x, com unboxing
		e.line("i64.const 0")
		if err := e.expr(n.Expr); err != nil {
			return err
		}
		e.line("i64.load offset=8")
		e.line("i64.sub")
		e.line("call $new_int")
	case *ast.Not:
		if err := e.expr(n.Expr); err != nil {
			return err
		}
		e.line("i32.load offset=8")
		e.line("i32.eqz")
		e.line("call $new_bool")
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
	case *ast.Dispatch:
		return e.dispatch(n)
	case *ast.StaticDispatch:
		return e.staticDispatch(n)
	default:
		return fmt.Errorf("expressao ainda nao suportada (%T)", x)
	}
	return nil
}

// object resolve um nome: self e o objeto atual, o resto e local de let,
// formal ou campo (leitura do proprio objeto pelo deslocamento do layout).
func (e *Emitter) object(n *ast.Object) error {
	if n.Name == "self" {
		e.line("local.get $self")
		return nil
	}
	if loc, ok := e.vars[n.Name]; ok {
		e.line("local.get %s", loc)
		return nil
	}
	if off := e.attrOffset(n.Name); off >= 0 {
		e.line("local.get $self")
		e.line("i32.load offset=%d", off)
		return nil
	}
	return fmt.Errorf("nome %s fora de alcance em %s", n.Name, e.class)
}

// newObj cria o objeto: basicas vao aos construtores do runtime, as do
// programa ao $C_new, e SELF_TYPE escolhe pela tag do self atual.
func (e *Emitter) newObj(n *ast.New) error {
	switch n.Type {
	case "Int":
		e.line("i64.const 0")
		e.line("call $new_int")
	case "Bool":
		e.line("i32.const 0")
		e.line("call $new_bool")
	case "String":
		e.line("call $new_str_empty")
	case "Object":
		e.line("i32.const %d", tagObject)
		e.line("i32.const 8")
		e.line("call $obj_new")
	case "IO":
		e.line("i32.const %d", tagIO)
		e.line("i32.const 8")
		e.line("call $obj_new")
	case "SELF_TYPE":
		return e.newSelf()
	default:
		if _, ok := e.sem.Tab.Classes[n.Type]; !ok {
			return fmt.Errorf("new de tipo inexistente %s", n.Type)
		}
		e.line("call $%s_new", n.Type)
	}
	return nil
}

// newSelf ramifica pela tag do self: cada subclasse constroi o seu.
// Despachar void cai no unreachable, que e erro de execucao como manda o manual.
func (e *Emitter) newSelf() error {
	e.line("(block $done (result i32)")
	e.level++
	for _, s := range e.subclasses(e.class) {
		e.line("local.get $self")
		e.line("i32.load offset=0")
		e.line("i32.const %d", e.tags[s])
		e.line("i32.eq")
		e.line("(if (then")
		e.level++
		e.line("call $%s_new", s)
		e.line("br $done")
		e.level--
		e.line("))")
	}
	e.line("unreachable")
	e.level--
	e.line(")")
	return nil
}

// voidCheck guarda o receptor no temporario e explode se for void: despacho
// sobre void e erro de execucao pelo manual.
func (e *Emitter) voidCheck() {
	e.line("local.set $tmp0")
	e.line("local.get $tmp0")
	e.line("i32.eqz")
	e.line("(if (then")
	e.level++
	e.line("unreachable")
	e.level--
	e.line("))")
}

// dispatch emite a chamada dinamica: avalia receptor e args, trava void, e
// ramifica pela tag ate a implementacao do dono (a mais especifica herdada).
// Os args vao para os temporarios para serem avaliados uma so vez.
func (e *Emitter) dispatch(n *ast.Dispatch) error {
	static := "SELF_TYPE"
	if n.Receiver == nil {
		e.line("local.get $self")
	} else {
		static = e.sem.Types[n.Receiver]
		if err := e.expr(n.Receiver); err != nil {
			return err
		}
	}
	if static == "SELF_TYPE" {
		static = e.class
	}
	if _, ok := e.sem.Tab.Methods[static][n.Method]; !ok {
		return fmt.Errorf("metodo %s inexistente em %s", n.Method, static)
	}
	e.voidCheck()
	for i := len(n.Args) - 1; i >= 0; i-- {
		e.line("local.set $a%d", i)
	}
	e.line("(block $done (result i32)")
	e.level++
	for _, s := range e.subclasses(static) {
		owner := e.sem.Tab.Methods[s][n.Method].Owner
		target, err := e.builtinTarget(owner, n.Method)
		if err != nil {
			return err
		}
		e.line("local.get $tmp0")
		e.line("i32.load offset=0")
		e.line("i32.const %d", e.tags[s])
		e.line("i32.eq")
		e.line("(if (then")
		e.level++
		e.line("local.get $tmp0")
		for i := range n.Args {
			e.line("local.get $a%d", i)
		}
		e.line("call %s", target)
		e.line("br $done")
		e.level--
		e.line("))")
	}
	e.line("unreachable")
	e.level--
	e.line(")")
	return nil
}

// staticDispatch emite a chamada com tipo estatico: trava void e chama direto
// a implementacao do dono, sem ramificar pela tag.
func (e *Emitter) staticDispatch(n *ast.StaticDispatch) error {
	if err := e.expr(n.Receiver); err != nil {
		return err
	}
	static := n.StaticType
	if static == "SELF_TYPE" {
		static = e.class
	}
	sig, ok := e.sem.Tab.Methods[static][n.Method]
	if !ok {
		return fmt.Errorf("metodo %s inexistente em %s", n.Method, static)
	}
	target, err := e.builtinTarget(sig.Owner, n.Method)
	if err != nil {
		return err
	}
	e.line("local.tee $tmp0")
	e.line("i32.eqz")
	e.line("(if (then")
	e.level++
	e.line("unreachable")
	e.level--
	e.line("))")
	e.line("local.get $tmp0")
	for _, a := range n.Args {
		if err := e.expr(a); err != nil {
			return err
		}
	}
	e.line("call %s", target)
	return nil
}

// builtinTarget resolve a funcao WAT do metodo do dono: abort e copy tem
// sinteticas no proprio emissor; type_name e passo 7 e o resto das basicas
// (IO, String) e passo 6.
func (e *Emitter) builtinTarget(owner, method string) (string, error) {
	if owner == "Object" && (method == "abort" || method == "copy") {
		return "$Object_" + method, nil
	}
	if info, ok := e.sem.Tab.Classes[owner]; ok && info.Node == nil {
		if method == "type_name" {
			return "", fmt.Errorf("type_name ainda nao suportado (passo 7)")
		}
		return "", fmt.Errorf("metodo %s de %s ainda nao suportado (passo 6)", method, owner)
	}
	return "$" + owner + "_" + method, nil
}

// unboxInt desempilha um endereco de Int e empilha o i64 de dentro.
func (e *Emitter) unboxInt() {
	e.line("i64.load offset=8")
}

// unboxBool desempilha um endereco de Bool e empilha o i32 de dentro.
func (e *Emitter) unboxBool() {
	e.line("i32.load offset=8")
}

// binOp emite os dois operandos, desempilha o direito para o temporario e
// opera com o esquerdo ja desembrulhado na ordem certa. O = delega inteiro
// ao equality, que escolhe a comparacao pelo tipo.
func (e *Emitter) binOp(n *ast.BinOp) error {
	if n.Op == "=" {
		return e.equality(n)
	}
	if err := e.expr(n.Left); err != nil {
		return err
	}
	if err := e.expr(n.Right); err != nil {
		return err
	}
	switch n.Op {
	case "+", "-", "*", "/", "<", "<=":
		e.line("local.set $tmp0")
		e.unboxInt()
		e.line("local.get $tmp0")
		e.unboxInt()
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
		}
		if n.Op == "<" || n.Op == "<=" {
			e.line("call $new_bool")
		} else {
			e.line("call $new_int")
		}
	default:
		return fmt.Errorf("operador ainda nao suportado (%q)", n.Op)
	}
	return nil
}

// equality: Int e Bool comparam valor desembrulhado; o resto (String,
// objetos, SELF_TYPE) compara endereco.
func (e *Emitter) equality(n *ast.BinOp) error {
	switch e.sem.Types[n.Left] {
	case "Int":
		if err := e.expr(n.Left); err != nil {
			return err
		}
		if err := e.expr(n.Right); err != nil {
			return err
		}
		e.line("local.set $tmp0")
		e.unboxInt()
		e.line("local.get $tmp0")
		e.unboxInt()
		e.line("i64.eq")
	case "Bool":
		if err := e.expr(n.Left); err != nil {
			return err
		}
		if err := e.expr(n.Right); err != nil {
			return err
		}
		e.line("local.set $tmp0")
		e.unboxBool()
		e.line("local.get $tmp0")
		e.unboxBool()
		e.line("i32.eq")
	default:
		if err := e.expr(n.Left); err != nil {
			return err
		}
		if err := e.expr(n.Right); err != nil {
			return err
		}
		e.line("i32.eq")
	}
	e.line("call $new_bool")
	return nil
}

// branch emite o if: a condicao desembrulha para i32 e os ramos ja sao
// enderecos do mesmo tipo WASM, entao o misturado funciona.
func (e *Emitter) branch(n *ast.If) error {
	if err := e.expr(n.Cond); err != nil {
		return err
	}
	e.unboxBool()
	e.line("(if (result i32)")
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

// loop emite o while como block+loop: desembrulha a condicao, sai quando e
// falsa, avalia o corpo e descarta o valor, e repete. O while tipa Object no
// COOL, entao o valor final e 0 (void).
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
	e.unboxBool()
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
// Tudo e endereco, entao init e declarado sempre combinam.
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

// defaultValue empilha o padrao encaixotado do COOL: Int 0, falso, String
// vazia, e void (endereco 0) para o resto.
func (e *Emitter) defaultValue(typ string) {
	switch typ {
	case "Int":
		e.line("i64.const 0")
		e.line("call $new_int")
	case "Bool":
		e.line("i32.const 0")
		e.line("call $new_bool")
	case "String":
		e.line("call $new_str_empty")
	default:
		e.line("i32.const 0")
	}
}

// assign avalia o valor e guarda no destino, mantendo uma copia na pilha:
// pela regra [ASSIGN] a atribuicao vale o proprio valor. Campo usa o
// temporario para reordenar endereco e valor para o store.
func (e *Emitter) assign(n *ast.Assign) error {
	if err := e.expr(n.Value); err != nil {
		return err
	}
	if loc, ok := e.vars[n.Target]; ok {
		e.line("local.tee %s", loc)
		return nil
	}
	if off := e.attrOffset(n.Target); off >= 0 {
		e.line("local.set $tmp0")
		e.line("local.get $self")
		e.line("local.get $tmp0")
		e.line("i32.store offset=%d", off)
		e.line("local.get $tmp0")
		return nil
	}
	return fmt.Errorf("atribuicao a %s fora de alcance em %s", n.Target, e.class)
}

// isVoid testa o endereco contra zero: 0 e void, resto nao e.
func (e *Emitter) isVoid(n *ast.IsVoid) error {
	if err := e.expr(n.Expr); err != nil {
		return err
	}
	e.line("i32.eqz")
	e.line("call $new_bool")
	return nil
}

// start monta a entrada do programa, que o manual define como (new Main).main().
// O resultado de main sai impresso com putint quando e Int (instrumentacao
// que some quando o out_int do IO estiver pronto) e descartado se nao e.
func (e *Emitter) start(prog *ast.Program) {
	e.b.WriteString("\n")
	e.level = 1
	e.line(`(func (export "_start")`)
	e.level = 2
	e.line("call $Main_new")
	e.line("call $Main_main")
	if mainReturnType(prog) == "Int" {
		e.line("i64.load offset=8")
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
