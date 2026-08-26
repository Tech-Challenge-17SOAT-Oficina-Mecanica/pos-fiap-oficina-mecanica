package ordemservico

import (
	"context"
	"testing"

	domainOrcamento "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type repositoryFake struct{}

func (repositoryFake) RegistrarItens(context.Context, RegistrarInput) (domainOrcamento.Resultado, error) {
	return domainOrcamento.Resultado{}, nil
}

func (repositoryFake) ConsultarOrcamentos(context.Context, string) ([]domainOrcamento.Resultado, error) {
	return nil, nil
}

func TestRegistrarItensRejeitaItemRepetido(t *testing.T) {
	useCase := NewRegistrarItens(repositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarInput{
		Tipo:  "PECA",
		Itens: []ItemInput{{ItemID: "item-1", Quantidade: 1}, {ItemID: "item-1", Quantidade: 2}},
	})
	if err != ErrItemRepetido {
		t.Fatalf("erro=%v, esperado %v", err, ErrItemRepetido)
	}
}

func TestRegistrarItensRejeitaQuantidadeFracionadaDePeca(t *testing.T) {
	useCase := NewRegistrarItens(repositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarInput{
		Tipo:  "PECA",
		Itens: []ItemInput{{ItemID: "item-1", Quantidade: 1.5}},
	})
	if err == nil {
		t.Fatal("quantidade fracionada aceita")
	}
}
