package ordemservico

import (
	"errors"
	"testing"
)

func TestAutorizacaoParaExecucao(t *testing.T) {
	for _, test := range []struct {
		name                string
		orcamento, servicos bool
		want                error
	}{
		{"sem autorizacoes", false, false, ErrOrcamentoNaoAprovado},
		{"sem orcamento", false, true, ErrOrcamentoNaoAprovado},
		{"sem servicos", true, false, ErrServicosNaoAutorizados},
		{"autorizado", true, true, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := ValidarAutorizacaoParaExecucao(test.orcamento, test.servicos); !errors.Is(got, test.want) {
				t.Fatalf("erro=%v esperado=%v", got, test.want)
			}
		})
	}
}

func TestRecursosParaExecucao(t *testing.T) {
	if err := ValidarRecursosParaExecucao(true); !errors.Is(err, ErrRecursosIndisponiveis) {
		t.Fatalf("recurso sem reserva: %v", err)
	}
	if err := ValidarRecursosParaExecucao(false); err != nil {
		t.Fatalf("recursos disponiveis: %v", err)
	}
}

func TestOrcamentosParaEntrega(t *testing.T) {
	if err := ValidarOrcamentosParaEntrega(0); !errors.Is(err, ErrValorFinalIndisponivel) {
		t.Fatalf("sem orcamento aprovado: %v", err)
	}
	for _, count := range []int{1, 2} {
		if err := ValidarOrcamentosParaEntrega(count); err != nil {
			t.Fatalf("orcamentos aprovados=%d erro=%v", count, err)
		}
	}
}
