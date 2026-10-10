package semantic

// Conforms diz se sub pode ser usado onde se espera sup
// Caminha de sub para cima pelos pais; achar sup no caminho prova tudo.
func (t *Table) Conforms(sub, sup string) bool {
	for cur := sub; cur != ""; cur = t.parentOf(cur) {
		if cur == sup {
			return true
		}
	}

	return false
}

// Join devolve o ancestral comum mais próximo (LCA) de a e b.
// Marca os ancestrais de a, depois sobre de b até o primeiro marcado.
func (t *Table) Join(a, b string) string {
	ancestors := map[string]bool{}
	for cur := a; cur != ""; cur = t.parentOf(cur) {
		ancestors[cur] = true
	}

	for cur := b; cur != ""; cur = t.parentOf(cur) {
		if ancestors[cur] {
			return cur
		}
	}

	return "Object" // inalcançável: toda caminhada termina em Object
}

// JoinAll dobra o Join sobre a lista (os Nramos do case)
func (t *Table) JoinAll(ts []string) string {
	j := ts[0]
	for _, x := range ts[1:] {
		j = t.Join(j, x)
	}
	return j
}
