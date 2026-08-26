package servico

import (
	"errors"
	"testing"
)

func TestNovo(t *testing.T) {
	t.Run("válido e normalizado", func(t *testing.T) {
		servico, err := Novo(NovoServicoInput{Nome: "  Tróca   de ÓLEO  ", Descricao: "  Completa  ", Valor: "0", TempoEstimadoMinutos: 60})
		if err != nil {
			t.Fatal(err)
		}
		if servico.Nome != "Tróca   de ÓLEO" || servico.NomeNormalizado != "troca de oleo" ||
			servico.Descricao != "Completa" || !servico.Ativo || servico.Version != 1 {
			t.Fatalf("serviço inesperado: %+v", servico)
		}
	})

	cases := []struct {
		name  string
		input NovoServicoInput
		err   error
	}{
		{"nome vazio", NovoServicoInput{Valor: "10", TempoEstimadoMinutos: 1}, ErrNomeObrigatorio},
		{"valor ausente", NovoServicoInput{Nome: "Teste", TempoEstimadoMinutos: 1}, ErrValorInvalido},
		{"valor negativo", NovoServicoInput{Nome: "Teste", Valor: "-0.01", TempoEstimadoMinutos: 1}, ErrValorInvalido},
		{"mais de duas casas", NovoServicoInput{Nome: "Teste", Valor: "10.001", TempoEstimadoMinutos: 1}, ErrValorInvalido},
		{"tempo ausente", NovoServicoInput{Nome: "Teste", Valor: "10"}, ErrTempoEstimadoInvalido},
		{"tempo negativo", NovoServicoInput{Nome: "Teste", Valor: "10", TempoEstimadoMinutos: -1}, ErrTempoEstimadoInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Novo(test.input)
			if !errors.Is(err, test.err) {
				t.Fatalf("erro %v, esperado %v", err, test.err)
			}
		})
	}
}

func TestNormalizarNome(t *testing.T) {
	if got := NormalizarNome("  Revisão\tBÁSICA com Limpeza  "); got != "revisao basica com limpeza" {
		t.Fatalf("normalização: %q", got)
	}
}
