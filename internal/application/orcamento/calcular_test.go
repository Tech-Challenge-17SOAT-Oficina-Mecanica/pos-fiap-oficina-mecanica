package orcamento

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

const (
	principalID    = "74000000-0000-0000-0000-000000000001"
	complementarID = "74000000-0000-0000-0000-000000000002"
	recusadoID     = "74000000-0000-0000-0000-000000000005"
	ordemID        = "70000000-0000-0000-0000-000000000001"
)

type calcularRepositoryFake struct {
	alvo        orcamento.Orcamento
	ordem       string
	erroBusca   error
	irmaos      []OrcamentoDaOS
	itensSalvos []orcamento.Item
	salvou      bool
}

func (fake *calcularRepositoryFake) BuscarParaCalculo(context.Context, string) (orcamento.Orcamento, string, error) {
	return fake.alvo, fake.ordem, fake.erroBusca
}

func (fake *calcularRepositoryFake) OrcamentosDaOrdem(context.Context, string) ([]OrcamentoDaOS, error) {
	return fake.irmaos, nil
}

func (fake *calcularRepositoryFake) SalvarItens(_ context.Context, _ string, itens []orcamento.Item) error {
	fake.salvou = true
	fake.itensSalvos = itens
	return nil
}

// Cenário do seed: principal CRIADO com serviço + peça + insumo, complementar CRIADO
// com um serviço, e um complementar RECUSADO que não pode entrar na conta.
func TestCalcularSomaPrincipalEComplementarIgnorandoRecusado(t *testing.T) {
	fake := &calcularRepositoryFake{
		ordem: ordemID,
		alvo: orcamento.Orcamento{
			ID: principalID, Tipo: orcamento.TipoPrincipal, Status: orcamento.StatusCriado,
			Itens: []orcamento.Item{
				{ID: "i1", Tipo: "SERVICO", Quantidade: 1, ValorUnitario: 150},
				{ID: "i2", Tipo: "PECA", Quantidade: 2, ValorUnitario: 45},
				{ID: "i3", Tipo: "INSUMO", Quantidade: 2, ValorUnitario: 32},
			},
		},
		irmaos: []OrcamentoDaOS{
			{ID: principalID, Status: orcamento.StatusCriado},
			{ID: complementarID, Status: orcamento.StatusCriado, Itens: []orcamento.Item{
				{ID: "i4", Tipo: "SERVICO", Quantidade: 1, ValorUnitario: 450},
			}},
			{ID: recusadoID, Status: orcamento.StatusRecusado, Itens: []orcamento.Item{
				{ID: "i5", Tipo: "PECA", Quantidade: 1, ValorUnitario: 180},
			}},
		},
	}

	resultado, err := NewCalcular(fake).Execute(context.Background(), principalID)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.ValorTotal != 304.00 {
		t.Fatalf("valorTotal = %.2f, esperado 304.00", resultado.ValorTotal)
	}
	if resultado.ValorTotalGeral != 754.00 {
		t.Fatalf("valorTotalGeral = %.2f, esperado 754.00 (304 + 450, sem os 180 do recusado)", resultado.ValorTotalGeral)
	}
	if !fake.salvou {
		t.Fatal("os itens recalculados deveriam ter sido salvos")
	}
}

func TestCalcularSemComplementarUsaSoOPrincipal(t *testing.T) {
	fake := &calcularRepositoryFake{
		ordem: ordemID,
		alvo: orcamento.Orcamento{
			ID: principalID, Tipo: orcamento.TipoPrincipal, Status: orcamento.StatusCriado,
			Itens: []orcamento.Item{{ID: "i1", Tipo: "SERVICO", Quantidade: 1, ValorUnitario: 150}},
		},
		irmaos: []OrcamentoDaOS{{ID: principalID, Status: orcamento.StatusCriado}},
	}

	resultado, err := NewCalcular(fake).Execute(context.Background(), principalID)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.ValorTotalGeral != 150.00 {
		t.Fatalf("valorTotalGeral = %.2f, esperado 150.00", resultado.ValorTotalGeral)
	}
}

// O alvo entra no total com os valores recém-calculados, não com os que estavam gravados.
func TestCalcularUsaValoresRecalculadosDoAlvo(t *testing.T) {
	fake := &calcularRepositoryFake{
		ordem: ordemID,
		alvo: orcamento.Orcamento{
			ID: principalID, Tipo: orcamento.TipoPrincipal, Status: orcamento.StatusCriado,
			Itens: []orcamento.Item{{ID: "i1", Quantidade: 2, ValorUnitario: 45, ValorTotal: 45}},
		},
		irmaos: []OrcamentoDaOS{
			{ID: principalID, Status: orcamento.StatusCriado, Itens: []orcamento.Item{
				{ID: "i1", Quantidade: 2, ValorUnitario: 45, ValorTotal: 45}, // valor velho
			}},
		},
	}

	resultado, err := NewCalcular(fake).Execute(context.Background(), principalID)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.ValorTotalGeral != 90.00 {
		t.Fatalf("valorTotalGeral = %.2f, esperado 90.00 (2 x 45), não o 45 gravado", resultado.ValorTotalGeral)
	}
}

func TestCalcularRejeitaIdentificadorInvalido(t *testing.T) {
	_, err := NewCalcular(&calcularRepositoryFake{}).Execute(context.Background(), "nao-e-uuid")
	if !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v", err)
	}
}

func TestCalcularPropagaRegrasDeDominio(t *testing.T) {
	casos := []struct {
		nome     string
		alvo     orcamento.Orcamento
		esperado error
	}{
		{"aprovado", orcamento.Orcamento{Tipo: orcamento.TipoPrincipal, Status: orcamento.StatusAprovado}, orcamento.ErrStatusNaoCalculavel},
		{"complementar sem principal", orcamento.Orcamento{Tipo: orcamento.TipoComplementar, Status: orcamento.StatusCriado}, orcamento.ErrComplementarSemPrincipal},
		{"sem itens", orcamento.Orcamento{Tipo: orcamento.TipoPrincipal, Status: orcamento.StatusCriado}, orcamento.ErrSemItens},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			fake := &calcularRepositoryFake{alvo: caso.alvo, ordem: ordemID}
			_, err := NewCalcular(fake).Execute(context.Background(), principalID)
			if !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
			if fake.salvou {
				t.Fatal("nada deveria ser salvo quando a validação falha")
			}
		})
	}
}
