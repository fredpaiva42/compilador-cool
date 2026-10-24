package codegen

import (
	"strings"
	"testing"

	"cool/ast"
	"cool/lexer"
	"cool/parser"
	"cool/semantic"
	"cool/token"
)

// parseAndCheck roda o pipeline do lexico ao semantico e devolve a AST e a analise.
func parseAndCheck(t *testing.T, src string) (*ast.Program, *semantic.Result) {
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
		t.Fatalf("parse: %v", err)
	}
	res, err := semantic.Analyze(prog)
	if err != nil {
		t.Fatalf("semantico: %v", err)
	}
	return prog, res
}

// generate roda o emissor inteiro sobre um programa COOL.
func generate(t *testing.T, src string) string {
	t.Helper()
	prog, res := parseAndCheck(t, src)
	wat, err := Generate(prog, res)
	if err != nil {
		t.Fatalf("gerar: %v", err)
	}
	return wat
}

// ops devolve so as instrucoes do corpo de Main.main, sem a declaracao da
// funcao nem o fechamento. O preludio do runtime nao entra: ele e fixo.
// Toda funcao declara (local $tmp0 i32), entao ele abre quase toda sequencia.
func ops(t *testing.T, src string) []string {
	t.Helper()
	wat := generate(t, src)
	i := strings.Index(wat, "(func $Main_main")
	if i < 0 {
		t.Fatalf("modulo sem Main.main:\n%s", wat)
	}
	j := strings.Index(wat[i:], "\n  )")
	if j < 0 {
		t.Fatalf("nao achei o fim de Main.main:\n%s", wat)
	}
	var out []string
	for _, ln := range strings.Split(wat[i:i+j], "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "(func") {
			continue
		}
		out = append(out, ln)
	}
	return out
}

// checkOps compara a sequencia de instrucoes geradas com a esperada.
func checkOps(t *testing.T, src string, want ...string) {
	t.Helper()
	got := ops(t, src)
	if len(got) != len(want) {
		t.Fatalf("instrucoes = %v, quer %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("instrucoes = %v, quer %v", got, want)
		}
	}
}

func TestMinimalModule(t *testing.T) {
	wat := generate(t, "class Main { main() : Int { 42 }; };")
	pieces := []string{
		"(module",
		"(func $Main_new (result i32)",
		"(func $Main_main (param $self i32) (result i32)",
		"i64.const 42",
		"call $new_int",
		`(func (export "_start")`,
		"call $Main_new",
		"call $Main_main",
		"i64.load offset=8",
		"call $putint",
		"call $proc_exit",
	}
	for _, piece := range pieces {
		if !strings.Contains(wat, piece) {
			t.Errorf("modulo sem %q\n--- modulo ---\n%s", piece, wat)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(wat), ")") {
		t.Error("modulo nao termina fechando o parentese")
	}
}

func TestArithmeticPrecedence(t *testing.T) {
	// 1 + 2 * 3: cada literal encaixota, cada operacao desembrulha e reencaixota
	checkOps(t, "class Main { main() : Int { 1 + 2 * 3 }; };",
		"(local $tmp0 i32)",
		"i64.const 1", "call $new_int",
		"i64.const 2", "call $new_int",
		"i64.const 3", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.mul", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.add", "call $new_int")
}

func TestArithmeticDivisionAndSubtraction(t *testing.T) {
	checkOps(t, "class Main { main() : Int { 100 / 2 - 8 }; };",
		"(local $tmp0 i32)",
		"i64.const 100", "call $new_int",
		"i64.const 2", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.div_s", "call $new_int",
		"i64.const 8", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.sub", "call $new_int")
}

func TestComparison(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { 1 < 2 }; };",
		"(local $tmp0 i32)",
		"i64.const 1", "call $new_int",
		"i64.const 2", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.lt_s", "call $new_bool")
	checkOps(t, "class Main { main() : Bool { 1 <= 2 }; };",
		"(local $tmp0 i32)",
		"i64.const 1", "call $new_int",
		"i64.const 2", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.le_s", "call $new_bool")
}

func TestNegUnboxes(t *testing.T) {
	checkOps(t, "class Main { main() : Int { ~5 }; };",
		"(local $tmp0 i32)",
		"i64.const 0",
		"i64.const 5", "call $new_int",
		"i64.load offset=8", "i64.sub", "call $new_int")
}

func TestNotUnboxes(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { not true }; };",
		"(local $tmp0 i32)",
		"i32.const 1", "call $new_bool",
		"i32.load offset=8", "i32.eqz", "call $new_bool")
}

func TestEqualityInt(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { 1 = 2 }; };",
		"(local $tmp0 i32)",
		"i64.const 1", "call $new_int",
		"i64.const 2", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.eq", "call $new_bool")
}

func TestEqualityBool(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { true = false }; };",
		"(local $tmp0 i32)",
		"i32.const 1", "call $new_bool",
		"i32.const 0", "call $new_bool",
		"local.set $tmp0", "i32.load offset=8", "local.get $tmp0", "i32.load offset=8",
		"i32.eq", "call $new_bool")
}

func TestEqualityObjects(t *testing.T) {
	// objetos comparam endereco, sem desembrulhar
	checkOps(t, "class Main { main() : Bool { let a : Object, b : Object in a = b }; };",
		"(local $tmp0 i32)", "(local $l0 i32)", "(local $l1 i32)",
		"i32.const 0", "local.set $l0",
		"i32.const 0", "local.set $l1",
		"local.get $l0", "local.get $l1", "i32.eq", "call $new_bool")
}

func TestBlockDropsAllButLast(t *testing.T) {
	checkOps(t, "class Main { main() : Int { { 1; 2; 3; } }; };",
		"(local $tmp0 i32)",
		"i64.const 1", "call $new_int", "drop",
		"i64.const 2", "call $new_int", "drop",
		"i64.const 3", "call $new_int")
}

func TestMainNonIntNotPrinted(t *testing.T) {
	wat := generate(t, "class Main { main() : Bool { true }; };")
	// o preludio contem a definicao de $putint, entao a checagem e no _start
	i := strings.Index(wat, `(func (export "_start")`)
	if i < 0 {
		t.Fatal("modulo sem _start")
	}
	if strings.Contains(wat[i:], "putint") {
		t.Error("resultado que nao e Int nao deve ser impresso")
	}
	if !strings.Contains(wat[i:], "drop") {
		t.Error("resultado descartado deveria aparecer como drop")
	}
}

func TestUnsupportedExpression(t *testing.T) {
	prog, res := parseAndCheck(t, `class Main { main() : Object { "oi" }; };`)
	_, err := Generate(prog, res)
	if err == nil {
		t.Fatal("esperava erro para expressao ainda nao suportada")
	}
	if !strings.Contains(err.Error(), "nao suportada") {
		t.Errorf("mensagem inesperada: %v", err)
	}
}

func TestDispatchCopyViaSelf(t *testing.T) {
	// copy() com receptor implicito despacha para $Object_copy
	wat := generate(t, "class Main { main() : Object { copy() }; };")
	for _, piece := range []string{"call $Object_copy", "(block $done (result i32)"} {
		if !strings.Contains(wat, piece) {
			t.Errorf("modulo sem %q", piece)
		}
	}
}

func TestClassTags(t *testing.T) {
	prog, _ := parseAndCheck(t, `class Zebra { main() : Object { self }; };
class Abacaxi { main() : Object { self }; };
class Main { main() : Object { self }; };`)
	tags := classTags(prog)
	want := map[string]int{
		"Object": 1, "IO": 2, "Int": 3, "String": 4, "Bool": 5,
		"Abacaxi": 6, "Main": 7, "Zebra": 8,
	}
	for name, tag := range want {
		if tags[name] != tag {
			t.Errorf("tag de %s = %d, quer %d", name, tags[name], tag)
		}
	}
}

func TestAttrLayout(t *testing.T) {
	_, res := parseAndCheck(t, `class A { x : Int; y : Bool; };
class B inherits A { z : Object; };
class Main { main() : Object { self }; };`)
	fields := attrLayout(res.Tab, "B")
	want := []attrField{{"x", "Int"}, {"y", "Bool"}, {"z", "Object"}}
	if len(fields) != len(want) {
		t.Fatalf("campos = %v, quer %v", fields, want)
	}
	for i := range want {
		if fields[i] != want[i] {
			t.Fatalf("campos = %v, quer %v", fields, want)
		}
	}
	e := &Emitter{sem: res, class: "B"}
	if off := e.attrOffset("y"); off != 16 {
		t.Errorf("deslocamento de y = %d, quer 16", off)
	}
	if off := e.attrOffset("w"); off != -1 {
		t.Errorf("deslocamento de w = %d, quer -1", off)
	}
}

func TestConstructorEmitted(t *testing.T) {
	wat := generate(t, "class Main { x : Int <- 5; main() : Int { x }; };")
	pieces := []string{
		"(func $Main_new (result i32)",
		"call $obj_new",
		"i64.const 5",
		"call $new_int",
		"i32.store offset=8",
		"local.get $self",
		"i32.load offset=8",
	}
	for _, piece := range pieces {
		if !strings.Contains(wat, piece) {
			t.Errorf("modulo sem %q\n--- modulo ---\n%s", piece, wat)
		}
	}
}

func TestNewSelfType(t *testing.T) {
	wat := generate(t, "class Main { main() : Object { new SELF_TYPE }; };")
	pieces := []string{
		"(block $done (result i32)",
		"call $Main_new",
		"unreachable",
	}
	for _, piece := range pieces {
		if !strings.Contains(wat, piece) {
			t.Errorf("modulo sem %q\n--- modulo ---\n%s", piece, wat)
		}
	}
}

func TestIfThenElse(t *testing.T) {
	checkOps(t, "class Main { main() : Int { if 1 < 2 then 10 else 20 fi }; };",
		"(local $tmp0 i32)",
		"i64.const 1", "call $new_int",
		"i64.const 2", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.lt_s", "call $new_bool",
		"i32.load offset=8",
		"(if (result i32)",
		"(then", "i64.const 10", "call $new_int", ")",
		"(else", "i64.const 20", "call $new_int", ")",
		")")
}

func TestIfMixedBranchesOk(t *testing.T) {
	// com tudo encaixotado, ramos de tipos diferentes sao so enderecos
	checkOps(t, "class Main { main() : Object { if true then 1 else false fi }; };",
		"(local $tmp0 i32)",
		"i32.const 1", "call $new_bool",
		"i32.load offset=8",
		"(if (result i32)",
		"(then", "i64.const 1", "call $new_int", ")",
		"(else", "i32.const 0", "call $new_bool", ")",
		")")
}

func TestLetAssign(t *testing.T) {
	checkOps(t, "class Main { main() : Int { let x : Int <- 40 in { x <- x + 2; } }; };",
		"(local $tmp0 i32)", "(local $l0 i32)",
		"i64.const 40", "call $new_int", "local.set $l0",
		"local.get $l0",
		"i64.const 2", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.add", "call $new_int",
		"local.tee $l0")
}

func TestLetDefault(t *testing.T) {
	checkOps(t, "class Main { main() : Int { let x : Int in x + 1 }; };",
		"(local $tmp0 i32)", "(local $l0 i32)",
		"i64.const 0", "call $new_int", "local.set $l0",
		"local.get $l0",
		"i64.const 1", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.add", "call $new_int")
}

func TestLetShadowing(t *testing.T) {
	// o interno ganha outro local e o externo volta a valer depois
	checkOps(t, "class Main { main() : Int { let x : Int <- 1 in { { let x : Int <- 2 in x; }; x; } }; };",
		"(local $tmp0 i32)", "(local $l0 i32)", "(local $l1 i32)",
		"i64.const 1", "call $new_int", "local.set $l0",
		"i64.const 2", "call $new_int", "local.set $l1",
		"local.get $l1", "drop",
		"local.get $l0")
}

func TestIsVoidBoxed(t *testing.T) {
	checkOps(t, "class Main { main() : Int { if isvoid 0 then 1 else 2 fi }; };",
		"(local $tmp0 i32)",
		"i64.const 0", "call $new_int",
		"i32.eqz", "call $new_bool",
		"i32.load offset=8",
		"(if (result i32)",
		"(then", "i64.const 1", "call $new_int", ")",
		"(else", "i64.const 2", "call $new_int", ")",
		")")
}

func TestIsVoidVoid(t *testing.T) {
	checkOps(t, "class Main { main() : Bool { let o : Object in isvoid o }; };",
		"(local $tmp0 i32)", "(local $l0 i32)",
		"i32.const 0", "local.set $l0",
		"local.get $l0", "i32.eqz", "call $new_bool")
}

func TestSelfIsParam(t *testing.T) {
	checkOps(t, "class Main { main() : Object { self }; };",
		"(local $tmp0 i32)",
		"local.get $self")
}

func TestWhileEmission(t *testing.T) {
	checkOps(t, "class Main { main() : Object { while true loop 1 pool }; };",
		"(local $tmp0 i32)",
		"(block $w0", "(loop $l0",
		"i32.const 1", "call $new_bool",
		"i32.load offset=8", "i32.eqz", "br_if $w0",
		"i64.const 1", "call $new_int", "drop", "br $l0",
		")", ")",
		"i32.const 0")
}

func TestAttrAssign(t *testing.T) {
	checkOps(t, "class Main { x : Int <- 1; main() : Int { { x <- x + 1; x; } }; };",
		"(local $tmp0 i32)",
		"local.get $self", "i32.load offset=8",
		"i64.const 1", "call $new_int",
		"local.set $tmp0", "i64.load offset=8", "local.get $tmp0", "i64.load offset=8",
		"i64.add", "call $new_int",
		"local.set $tmp0", "local.get $self", "local.get $tmp0",
		"i32.store offset=8", "local.get $tmp0",
		"drop",
		"local.get $self", "i32.load offset=8")
}

func TestDynamicDispatch(t *testing.T) {
	// a.tag 7 (Cachorro) chama o override; Gato herdaria o de Animal
	checkOps(t, `class Animal { valor() : Int { 1 }; };
class Cachorro inherits Animal { valor() : Int { 2 }; };
class Gato inherits Animal { };
class Main { main() : Int { let a : Animal <- new Cachorro in a.valor() }; };`,
		"(local $tmp0 i32)", "(local $l0 i32)",
		"call $Cachorro_new", "local.set $l0",
		"local.get $l0",
		"local.set $tmp0", "local.get $tmp0", "i32.eqz",
		"(if (then", "unreachable", "))",
		"(block $done (result i32)",
		"local.get $tmp0", "i32.load offset=0", "i32.const 6", "i32.eq",
		"(if (then", "local.get $tmp0", "call $Animal_valor", "br $done", "))",
		"local.get $tmp0", "i32.load offset=0", "i32.const 7", "i32.eq",
		"(if (then", "local.get $tmp0", "call $Cachorro_valor", "br $done", "))",
		"local.get $tmp0", "i32.load offset=0", "i32.const 8", "i32.eq",
		"(if (then", "local.get $tmp0", "call $Animal_valor", "br $done", "))",
		"unreachable", ")")
}

func TestSelfDispatch(t *testing.T) {
	checkOps(t, "class Main { f() : Int { 3 }; main() : Int { f() }; };",
		"(local $tmp0 i32)",
		"local.get $self",
		"local.set $tmp0", "local.get $tmp0", "i32.eqz",
		"(if (then", "unreachable", "))",
		"(block $done (result i32)",
		"local.get $tmp0", "i32.load offset=0", "i32.const 6", "i32.eq",
		"(if (then", "local.get $tmp0", "call $Main_f", "br $done", "))",
		"unreachable", ")")
}

func TestStaticDispatch(t *testing.T) {
	checkOps(t, `class A { f() : Int { 7 }; };
class B inherits A { };
class Main { main() : Int { (new B)@A.f() }; };`,
		"(local $tmp0 i32)",
		"call $B_new",
		"local.tee $tmp0", "i32.eqz",
		"(if (then", "unreachable", "))",
		"local.get $tmp0", "call $A_f")
}

func TestBuiltinCopy(t *testing.T) {
	// copy de Int passa pelo $Object_copy sintetizado
	checkOps(t, "class Main { main() : Int { let c : Int <- 5 in c.copy() }; };",
		"(local $tmp0 i32)", "(local $l0 i32)",
		"i64.const 5", "call $new_int", "local.set $l0",
		"local.get $l0",
		"local.set $tmp0", "local.get $tmp0", "i32.eqz",
		"(if (then", "unreachable", "))",
		"(block $done (result i32)",
		"local.get $tmp0", "i32.load offset=0", "i32.const 3", "i32.eq",
		"(if (then", "local.get $tmp0", "call $Object_copy", "br $done", "))",
		"unreachable", ")")
}

func TestStringMethodNotYet(t *testing.T) {
	prog, res := parseAndCheck(t, "class Main { main() : Object { (new String).length() }; };")
	_, err := Generate(prog, res)
	if err == nil {
		t.Fatal("esperava erro para metodo de String")
	}
	if !strings.Contains(err.Error(), "passo 6") {
		t.Errorf("mensagem inesperada: %v", err)
	}
}

func TestTypeNameNotYet(t *testing.T) {
	prog, res := parseAndCheck(t, "class Main { main() : Object { 1.type_name() }; };")
	_, err := Generate(prog, res)
	if err == nil {
		t.Fatal("esperava erro para type_name")
	}
	if !strings.Contains(err.Error(), "passo 7") {
		t.Errorf("mensagem inesperada: %v", err)
	}
}
