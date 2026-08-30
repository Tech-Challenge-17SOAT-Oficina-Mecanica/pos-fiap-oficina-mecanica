package estoque

import (
	"errors"
	"fmt"
	"testing"
)

func TestNovaSaidaCadastro(t *testing.T) {
	item := ItemSaida{ItemID: "90000000-0000-0000-0000-000000000001", Quantidade: 1}

	saida, err := NovaSaidaCadastro("80000000-0000-0000-0000-000000000001", []ItemSaida{item}, true)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if saida.OrdemServicoID == "" || !saida.LiberarSaldoNaoUsado || len(saida.Itens) != 1 {
		t.Fatalf("saida invalida: %+v", saida)
	}
}

func TestNovaSaidaCadastroErros(t *testing.T) {
	item := ItemSaida{ItemID: "90000000-0000-0000-0000-000000000001", Quantidade: 1}
	tests := []struct {
		nome         string
		itens        []ItemSaida
		querErro     error
		querMensagem string
	}{
		{"itens obrigatorios", nil, ErrItensObrigatorios, ""},
		{"item repetido", []ItemSaida{item, item}, ErrItemRepetido, ""},
		{"quantidade invalida", []ItemSaida{{ItemID: item.ItemID, Quantidade: 0}}, nil, "quantidade deve ser maior que zero"},
	}

	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			_, err := NovaSaidaCadastro("80000000-0000-0000-0000-000000000001", test.itens, true)
			if test.querMensagem != "" {
				if err == nil || err.Error() != test.querMensagem {
					t.Fatalf("erro=%v, quer mensagem=%q", err, test.querMensagem)
				}
				return
			}
			if !errors.Is(err, test.querErro) {
				t.Fatalf("erro=%v, quer=%v", err, test.querErro)
			}
		})
	}
}

func TestNovaSaidaCadastroLimiteDeItens(t *testing.T) {
	itens := make([]ItemSaida, 201)
	for i := range itens {
		itens[i] = ItemSaida{ItemID: fmt.Sprintf("item-%d", i), Quantidade: 1}
	}

	_, err := NovaSaidaCadastro("80000000-0000-0000-0000-000000000001", itens, true)
	if !errors.Is(err, ErrItensExcedemLimite) {
		t.Fatalf("erro=%v", err)
	}
}
