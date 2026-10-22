package semantic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// checkArquivo lê um .cl e roda o pipeline inteiro sobre ele.
func checkArquivo(t *testing.T, nome string) error {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("..", "examples", nome))
	if err != nil {
		t.Fatalf("ler %s: %v", nome, err)
	}
	return checkSrc(t, string(src))
}

// TestExemplosSemanticos roda os exemplos que exercitam este pacote:
// os sem_ok_ precisam passar, os sem_erro_ precisam falhar com a
// mensagem esperada.
func TestExemplosSemanticos(t *testing.T) {
	casos := []struct {
		arquivo string
		erro    string // vazio = o programa deve passar
	}{
		{"sem_ok_heranca.cl", ""},
		{"sem_ok_selftype.cl", ""},
		{"sem_ok_escopos.cl", ""},
		{"sem_erro_override.cl", "com assinatura diferente"},
		{"sem_erro_atributo.cl", "redefine herdado"},
		{"sem_erro_dispatch.cl", "inexistente em Animal"},
	}

	for _, c := range casos {
		t.Run(c.arquivo, func(t *testing.T) {
			err := checkArquivo(t, c.arquivo)
			if c.erro == "" {
				if err != nil {
					t.Fatalf("deveria passar: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("deveria falhar com %q", c.erro)
			}
			if !strings.Contains(err.Error(), c.erro) {
				t.Fatalf("mensagem %q não contém %q", err.Error(), c.erro)
			}
		})
	}
}
