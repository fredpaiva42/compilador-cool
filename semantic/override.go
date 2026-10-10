package semantic

import "fmt"

// signaturesEqual diz se duas assinaturas são idênticas para fins de
// override: mesma aridade, mesmos tipos de formais, mesmo retorno.
// Os NOMES dos formais não entram: chamada é por posição, não por nome.
func signaturesEqual(a, b *MethodSig) bool {
	if len(a.ParamTypes) != len(b.ParamTypes) {
		return false
	}

	for i := range a.ParamTypes {
		if a.ParamTypes[i] != b.ParamTypes[i] {
			return false
		}
	}

	return a.Return == b.Return
}

func (t *Table) parentOf(name string) string {
	if name == "Object" {
		return ""
	}

	if p := t.Classes[name].Parent; p != "" {
		return p
	}

	return "Object"
}

// checkOverride compara o método que C declara com cada ancestral que
// também declara o mesmo nome. A primeira assinatura encontrada subindo
// é o contrato a respeitar.
func (t *Table) checkOverride(className, methodName string, sig *MethodSig) error {
	for cur := t.parentOf(className); cur != ""; cur = t.parentOf(cur) {
		if parentSig, ok := t.Methods[cur][methodName]; ok {
			if !signaturesEqual(sig, parentSig) {
				return fmt.Errorf("método %s de %s redefine %s.%s com assinatura diferente (linha %d)",
					methodName, className, parentSig.Owner, methodName, sig.Line)
			}
			return nil
		}
	}
	return nil
}
