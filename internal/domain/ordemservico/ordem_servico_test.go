package ordemservico

import (
	"errors"
	"testing"
)

func TestNovoProblemaRelatado(t *testing.T) {
	problema, err := NovoProblemaRelatado("  Ruído ao frear  ", "  Há uma semana  ")
	if err != nil || problema.Descricao != "Ruído ao frear" || problema.Observacoes != "Há uma semana" {
		t.Fatalf("problema = %+v, erro = %v", problema, err)
	}
}

func TestNovoProblemaRelatadoSemObservacoes(t *testing.T) {
	problema, err := NovoProblemaRelatado("Ruído", "")
	if err != nil || problema.Observacoes != "" {
		t.Fatalf("problema = %+v, erro = %v", problema, err)
	}
}

func TestNovoProblemaRelatadoRejeitaDescricaoVazia(t *testing.T) {
	_, err := NovoProblemaRelatado("  ", "observação")
	if !errors.Is(err, ErrDescricaoObrigatoria) {
		t.Fatalf("erro = %v", err)
	}
}
