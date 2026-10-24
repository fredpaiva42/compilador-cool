// O arquivo rt.wat e embutido no binario na compilacao (go:embed), entao o
// compilador nao depende de arquivos externos em tempo de execucao. O trecho
// entre as marcas ">>> RUNTIME" e "<<< RUNTIME" e o que o emissor cola no
// inicio de todo modulo gerado: imports, memoria, globais e as funcoes do
// runtime. O que vem depois das marcas (dados e _start) pertence a bateria
// de teste e nao entra no codigo gerado.
package runtime

import (
	_ "embed"
	"strings"
)

//go:embed rt.wat
var source string

const (
	startMark = ";; >>> RUNTIME"
	endMark   = ";; <<< RUNTIME"
)

// Prelude devolve o trecho do rt.wat que compoe o inicio de todo modulo
// gerado. Entra em panico se as marcas nao estiverem no arquivo, porque isso
// e erro de manutencao do runtime, nao erro de programa do usuario.
func Prelude() string {
	i := strings.Index(source, startMark)
	if i < 0 {
		panic("runtime: rt.wat sem a marca " + startMark)
	}
	rest := source[i:]
	j := strings.Index(rest, endMark)
	if j < 0 {
		panic("runtime: rt.wat sem a marca " + endMark)
	}
	return rest[:j]
}
