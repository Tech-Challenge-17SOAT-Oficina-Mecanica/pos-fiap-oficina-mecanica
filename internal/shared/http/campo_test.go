package http

import (
	"errors"
	"net/url"
	"testing"
)

func TestLerBooleano(t *testing.T) {
	casos := []struct {
		valor    string
		esperado bool
		erro     bool
	}{
		{"", false, false},
		{"true", true, false},
		{"false", false, false},
		{"abc", false, true},
		{"TRUE", false, true},
		{"1", false, true},
		{"sim", false, true},
	}

	for _, caso := range casos {
		t.Run("valor="+caso.valor, func(t *testing.T) {
			query := url.Values{}
			if caso.valor != "" {
				query.Set("incluirInativos", caso.valor)
			}

			obtido, err := LerBooleano(query, "incluirInativos")
			if caso.erro {
				if err == nil {
					t.Fatalf("valor %q deveria ser rejeitado, não virar false silenciosamente", caso.valor)
				}
				if campo := CampoDoErro(err); campo != "incluirInativos" {
					t.Fatalf("campo do erro = %q", campo)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if obtido != caso.esperado {
				t.Fatalf("obtido = %v, esperado %v", obtido, caso.esperado)
			}
		})
	}
}

func TestCampoDoErroPaginacao(t *testing.T) {
	if campo := CampoDoErro(ErrTamanhoInvalido); campo != "tamanho" {
		t.Fatalf("campo = %q, esperado tamanho", campo)
	}
	if campo := CampoDoErro(ErrPaginaInvalida); campo != "pagina" {
		t.Fatalf("campo = %q, esperado pagina", campo)
	}
	if campo := CampoDoErro(errors.New("erro simples")); campo != "" {
		t.Fatalf("erro sem campo deveria devolver vazio, veio %q", campo)
	}
}
