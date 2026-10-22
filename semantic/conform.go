package semantic

// Conforms informa se um valor de tipo sub pode ser usado onde se espera sup.
func (t *Table) Conforms(sub, sup string) bool {
	for cur := sub; cur != ""; cur = t.parentOf(cur) {
		if cur == sup {
			return true
		}
	}

	return false
}

// Join devolve o ancestral comum mais próximo (LCA) de a e b.
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

func (t *Table) JoinAll(ts []string) string {
	j := ts[0]
	for _, x := range ts[1:] {
		j = t.Join(j, x)
	}
	return j
}
