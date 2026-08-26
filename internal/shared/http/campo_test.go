package http

import (
	"errors"
	"testing"
)

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
