package runtime

import (
	"strings"
	"testing"
)

func TestPreludioTemAsPecas(t *testing.T) {
	p := Preludio()
	pecas := []string{
		"(module",
		`(import "wasi_snapshot_preview1" "fd_write"`,
		`(memory (export "memory")`,
		"(global $heap",
		"(func $putstr",
		"(func $alloc",
		"(func $putint",
		"(func $read_line",
		"(func $read_int",
	}
	for _, peca := range pecas {
		if !strings.Contains(p, peca) {
			t.Errorf("preludio nao contem %q", peca)
		}
	}
}

func TestPreludioNaoTemABateria(t *testing.T) {
	p := Preludio()
	proibidas := []string{"(data", "_start", "alloc FALHOU"}
	for _, proibida := range proibidas {
		if strings.Contains(p, proibida) {
			t.Errorf("preludio nao deveria conter %q", proibida)
		}
	}
}

func TestPreludioDeixaOModuloAberto(t *testing.T) {
	// o preludio abre o modulo e nao o fecha: quem fecha e o emissor,
	// depois de colar os dados e as funcoes geradas
	p := Preludio()
	abre := strings.Count(p, "(")
	fecha := strings.Count(p, ")")
	if abre != fecha+1 {
		t.Errorf("esperava exatamente um parentese aberto a mais (abre %d, fecha %d)", abre, fecha)
	}
}
