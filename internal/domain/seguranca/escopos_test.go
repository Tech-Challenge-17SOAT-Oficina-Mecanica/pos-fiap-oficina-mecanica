package seguranca

import (
	"slices"
	"strings"
	"testing"
)

func TestEscopoValido(t *testing.T) {
	for _, escopo := range EscoposOficiais {
		if !EscopoValido(escopo) {
			t.Fatalf("escopo oficial %q foi rejeitado", escopo)
		}
	}
	for _, invalido := range []string{"", "estoque", "estoque:apagar", "ESTOQUE:LER", " estoque:ler"} {
		if EscopoValido(invalido) {
			t.Fatalf("escopo %q não deveria ser aceito", invalido)
		}
	}
}

func TestEscoposOficiaisSemDuplicata(t *testing.T) {
	vistos := map[string]bool{}
	for _, escopo := range EscoposOficiais {
		if vistos[escopo] {
			t.Fatalf("escopo %q aparece duas vezes na lista", escopo)
		}
		vistos[escopo] = true
	}
}

// O formato importa: o middleware compara o escopo por igualdade exata, então um
// "Estoque:Ler" na lista viraria um 403 silencioso.
func TestFormatoDosEscopos(t *testing.T) {
	for _, escopo := range EscoposOficiais {
		recurso, acao, encontrou := strings.Cut(escopo, ":")
		if !encontrou {
			t.Fatalf("escopo %q não segue o formato recurso:acao", escopo)
		}
		if recurso == "" || acao == "" {
			t.Fatalf("escopo %q tem parte vazia", escopo)
		}
		if escopo != strings.ToLower(escopo) {
			t.Fatalf("escopo %q deveria ser todo minúsculo", escopo)
		}
		if !slices.Contains([]string{"ler", "escrever", "decidir", "movimentar"}, acao) {
			t.Fatalf("escopo %q usa ação inesperada %q", escopo, acao)
		}
	}
}
