package runtime

import (
	"strings"
	"testing"
)

func TestPreludeHasPieces(t *testing.T) {
	p := Prelude()
	pieces := []string{
		"(module",
		`(import "wasi_snapshot_preview1" "fd_write"`,
		`(memory (export "memory")`,
		"(global $heap",
		"(func $putstr",
		"(func $alloc",
		"(func $putint",
		"(func $read_line",
		"(func $read_int",
		"(func $obj_new",
		"(func $new_int",
		"(func $new_bool",
		"(func $new_str_empty",
		"(func $copy_obj",
		"(func $abort",
		"(func $str_concat",
		"(func $str_substr",
		"(func $str_eq",
		"(func $io_out_string",
		"(func $io_out_int",
		"(func $io_in_string",
		"(func $io_in_int",
	}
	for _, piece := range pieces {
		if !strings.Contains(p, piece) {
			t.Errorf("preludio nao contem %q", piece)
		}
	}
}

func TestPreludeHasNoBattery(t *testing.T) {
	p := Prelude()
	forbidden := []string{"(data", "_start", "alloc FALHOU"}
	for _, f := range forbidden {
		if strings.Contains(p, f) {
			t.Errorf("preludio nao deveria conter %q", f)
		}
	}
}

func TestPreludeLeavesModuleOpen(t *testing.T) {
	// o preludio abre o modulo e nao o fecha: quem fecha e o emissor,
	// depois de colar os dados e as funcoes geradas
	p := Prelude()
	open := strings.Count(p, "(")
	closed := strings.Count(p, ")")
	if open != closed+1 {
		t.Errorf("esperava exatamente um parentese aberto a mais (abre %d, fecha %d)", open, closed)
	}
}
