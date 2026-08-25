package orcamento

import (
	"errors"
	"testing"
)

func inteiro(value int) *int { return &value }

func calculoPrincipalValido() Calculo {
	return Calculo{
		Tipo: TipoPrincipal, Status: StatusCriado, CapacidadeDiariaOS: 3, MinutosProdutivosDia: 480, OrdensNaFila: 4,
		Itens: []Item{
			{Tipo: "SERVICO", Quantidade: "1.000", ValorUnitario: "150.00", TempoServicoMinutos: 500, MaterialDisponivel: true},
			{Tipo: "PECA", Quantidade: "1.000", ValorUnitario: "50.00", MaterialDisponivel: false, PrazoEntregaDias: inteiro(5)},
		},
	}
}

func TestEstimativaOrcamentoPrincipal(t *testing.T) {
	got, err := calculoPrincipalValido().EstimativaEntregaDias()
	// 5 dias de material + 2 dias de serviço + 2 dias de fila.
	if err != nil || got != 9 {
		t.Fatalf("estimativa=%d erro=%v", got, err)
	}
}

func TestEstimativaOrcamentoComplementar(t *testing.T) {
	principal := 4
	calculo := Calculo{
		Tipo: TipoComplementar, Status: StatusCriado, OrcamentoOriginalID: "principal", OrcamentoPrincipalID: "principal",
		EstimativaPrincipalDias: &principal, CapacidadeDiariaOS: 3, MinutosProdutivosDia: 480, OrdensNaFila: 99,
		Itens: []Item{{Tipo: "SERVICO", Quantidade: "1", ValorUnitario: "10.00", TempoServicoMinutos: 60, MaterialDisponivel: true}},
	}
	got, err := calculo.EstimativaEntregaDias()
	if err != nil || got != 5 {
		t.Fatalf("estimativa=%d erro=%v", got, err)
	}
}

func TestValidacoesCalculo(t *testing.T) {
	base := calculoPrincipalValido()
	cases := []struct {
		name string
		edit func(*Calculo)
		want error
	}{
		{"status", func(c *Calculo) { c.Status = "APROVADO" }, ErrStatusInvalido},
		{"tipo", func(c *Calculo) { c.Tipo = "INVALIDO" }, ErrTipoInvalido},
		{"principal com original", func(c *Calculo) { c.OrcamentoOriginalID = "id" }, ErrVinculoPrincipalInvalido},
		{"sem itens", func(c *Calculo) { c.Itens = nil }, ErrSemItens},
		{"quantidade", func(c *Calculo) { c.Itens[0].Quantidade = "0" }, ErrItemInvalido},
		{"valor", func(c *Calculo) { c.Itens[0].ValorUnitario = "-1" }, ErrItemInvalido},
		{"tipo item", func(c *Calculo) { c.Itens[0].Tipo = "OUTRO" }, ErrItemInvalido},
		{"tempo", func(c *Calculo) { c.Itens[0].TempoServicoMinutos = 0 }, ErrTempoServicoIndisponivel},
		{"prazo", func(c *Calculo) { c.Itens[1].PrazoEntregaDias = nil }, ErrPrazoIndisponivel},
		{"capacidade", func(c *Calculo) { c.CapacidadeDiariaOS = 0 }, ErrConfiguracaoInvalida},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			calculo := base
			calculo.Itens = append([]Item(nil), base.Itens...)
			test.edit(&calculo)
			_, err := calculo.EstimativaEntregaDias()
			if !errors.Is(err, test.want) {
				t.Fatalf("erro=%v esperado=%v", err, test.want)
			}
		})
	}
}

func TestComplementarSemPrincipal(t *testing.T) {
	calculo := calculoPrincipalValido()
	calculo.Tipo = TipoComplementar
	_, err := calculo.EstimativaEntregaDias()
	if !errors.Is(err, ErrVinculoPrincipalInvalido) {
		t.Fatalf("erro=%v", err)
	}
}
