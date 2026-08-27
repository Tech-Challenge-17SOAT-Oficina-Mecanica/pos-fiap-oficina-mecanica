package orcamento

import (
	"errors"
	"testing"
)

func TestRecalcularAplicaQuantidadeVezesValor(t *testing.T) {
	itens, total, err := Recalcular([]Item{
		{Tipo: "SERVICO", Quantidade: 1, ValorUnitario: 150.00, ValorTotal: 0},
		{Tipo: "PECA", Quantidade: 2, ValorUnitario: 45.00, ValorTotal: 45.00}, // desatualizado
		{Tipo: "INSUMO", Quantidade: 2, ValorUnitario: 32.00, ValorTotal: 64.00},
	})
	if err != nil {
		t.Fatal(err)
	}
	if itens[1].ValorTotal != 90.00 {
		t.Fatalf("valorTotal do item = %.2f, esperado 90.00 (o gravado estava errado)", itens[1].ValorTotal)
	}
	if total != 304.00 {
		t.Fatalf("total = %.2f, esperado 304.00", total)
	}
}

func TestRecalcularAceitaQuantidadeFracionada(t *testing.T) {
	_, total, err := Recalcular([]Item{{Tipo: "INSUMO", Quantidade: 2.5, ValorUnitario: 32.00}})
	if err != nil {
		t.Fatal(err)
	}
	if total != 80.00 {
		t.Fatalf("total = %.2f, esperado 80.00", total)
	}
}

func TestRecalcularRejeitaEntradaInvalida(t *testing.T) {
	casos := []struct {
		nome     string
		itens    []Item
		esperado error
	}{
		{"sem itens", nil, ErrSemItens},
		{"lista vazia", []Item{}, ErrSemItens},
		{"quantidade zero", []Item{{Quantidade: 0, ValorUnitario: 10}}, ErrItemInvalido},
		{"quantidade negativa", []Item{{Quantidade: -1, ValorUnitario: 10}}, ErrItemInvalido},
		{"valor negativo", []Item{{Quantidade: 1, ValorUnitario: -1}}, ErrItemInvalido},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if _, _, err := Recalcular(caso.itens); !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
		})
	}
}

// RF-ORC-05 e RF-ORC-06: criado e aprovado somam no total geral, recusado não.
func TestEntraNoTotalGeral(t *testing.T) {
	casos := map[string]bool{
		StatusCriado:   true,
		StatusAprovado: true,
		StatusRecusado: false,
	}
	for status, esperado := range casos {
		t.Run(status, func(t *testing.T) {
			if obtido := (Orcamento{Status: status}).EntraNoTotalGeral(); obtido != esperado {
				t.Fatalf("status %q entra no total = %v, esperado %v", status, obtido, esperado)
			}
		})
	}
}

// Só CRIADO pode ser calculado: recalcular um aprovado mudaria o valor de algo que o
// cliente já respondeu. E o complementar precisa apontar para o PRINCIPAL da mesma OS —
// os vínculos recebidos são, por construção, apenas os orçamentos daquela OS.
func TestValidarParaCalculo(t *testing.T) {
	principalDaOS := []Vinculo{
		{ID: "principal-1", Tipo: TipoPrincipal},
		{ID: "complementar-1", Tipo: TipoComplementar},
	}

	casos := []struct {
		nome      string
		orcamento Orcamento
		vinculos  []Vinculo
		esperado  error
	}{
		{"principal criado", Orcamento{Tipo: TipoPrincipal, Status: StatusCriado}, principalDaOS, nil},
		{"complementar vinculado ao principal", Orcamento{Tipo: TipoComplementar, Status: StatusCriado, OriginalID: "principal-1"}, principalDaOS, nil},
		{"aprovado", Orcamento{Tipo: TipoPrincipal, Status: StatusAprovado}, principalDaOS, ErrStatusNaoCalculavel},
		{"recusado", Orcamento{Tipo: TipoPrincipal, Status: StatusRecusado}, principalDaOS, ErrStatusNaoCalculavel},
		{"complementar sem vinculo", Orcamento{Tipo: TipoComplementar, Status: StatusCriado}, principalDaOS, ErrComplementarSemPrincipal},
		{"vinculo de outra OS", Orcamento{Tipo: TipoComplementar, Status: StatusCriado, OriginalID: "principal-de-outra-os"}, principalDaOS, ErrVinculoInvalido},
		{"vinculo aponta para outro complementar", Orcamento{Tipo: TipoComplementar, Status: StatusCriado, OriginalID: "complementar-1"}, principalDaOS, ErrVinculoInvalido},
		{"OS sem nenhum principal", Orcamento{Tipo: TipoComplementar, Status: StatusCriado, OriginalID: "principal-1"}, []Vinculo{{ID: "complementar-1", Tipo: TipoComplementar}}, ErrVinculoInvalido},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if err := caso.orcamento.ValidarParaCalculo(caso.vinculos); !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
		})
	}
}

func TestArredondamentoDuasCasas(t *testing.T) {
	_, total, err := Recalcular([]Item{{Quantidade: 3, ValorUnitario: 33.333}})
	if err != nil {
		t.Fatal(err)
	}
	if total != 100.00 {
		t.Fatalf("total = %.4f, esperado 100.00", total)
	}
}
