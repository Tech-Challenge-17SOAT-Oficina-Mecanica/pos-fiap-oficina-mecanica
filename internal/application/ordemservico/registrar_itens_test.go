package ordemservico

import (
	"context"
	"testing"

	domainOrcamento "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type registrarItensRepositoryFake struct{}

func (registrarItensRepositoryFake) RegistrarItens(context.Context, RegistrarInput) (domainOrcamento.Resultado, error) {
	return domainOrcamento.Resultado{}, nil
}

func TestRegistrarItensRejeitaItemRepetido(t *testing.T) {
	useCase := NewRegistrarItens(registrarItensRepositoryFake{})
	const itemID = "10000000-0000-0000-0000-000000000001"
	_, err := useCase.Execute(context.Background(), RegistrarInput{
		Tipo:  "PECA",
		Itens: []ItemInput{{ItemID: itemID, Quantidade: 1}, {ItemID: itemID, Quantidade: 2}},
	})
	if err != ErrItemRepetido {
		t.Fatalf("erro=%v, esperado %v", err, ErrItemRepetido)
	}
}

func TestRegistrarItensRejeitaQuantidadeFracionadaDePeca(t *testing.T) {
	useCase := NewRegistrarItens(registrarItensRepositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarInput{
		Tipo:  "PECA",
		Itens: []ItemInput{{ItemID: "10000000-0000-0000-0000-000000000001", Quantidade: 1.5}},
	})
	if err == nil {
		t.Fatal("quantidade fracionada aceita")
	}
}

func TestRegistrarItensRejeitaItemIDInvalido(t *testing.T) {
	useCase := NewRegistrarItens(registrarItensRepositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarInput{
		Tipo:  "PECA",
		Itens: []ItemInput{{ItemID: "item-1", Quantidade: 1}},
	})
	if err != ErrItemIDInvalido {
		t.Fatalf("erro=%v, esperado %v", err, ErrItemIDInvalido)
	}
}
