package estoque

import (
	"errors"
	"fmt"
	"testing"
)

func TestNovaEntradaCadastro(t *testing.T) {
	item := ItemEntrada{
		ItemID:        "90000000-0000-0000-0000-000000000001",
		Quantidade:    1,
		CustoUnitario: 10,
	}

	entrada, err := NovaEntradaCadastro(" NF-123 ", "90000000-0000-0000-0000-000000000002", "pedido-1", true, []ItemEntrada{item})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if entrada.DocumentoOrigem != "NF-123" {
		t.Fatalf("documentoOrigem=%q", entrada.DocumentoOrigem)
	}
	if entrada.FornecedorID != "90000000-0000-0000-0000-000000000002" || entrada.PedidoCompraID != "pedido-1" || !entrada.ConfirmarDivergencia {
		t.Fatalf("entrada invalida: %+v", entrada)
	}
}

func TestNovaEntradaCadastroErros(t *testing.T) {
	item := ItemEntrada{
		ItemID:        "90000000-0000-0000-0000-000000000001",
		Quantidade:    1,
		CustoUnitario: 10,
	}
	tests := []struct {
		nome            string
		documentoOrigem string
		fornecedorID    string
		itens           []ItemEntrada
		querErro        error
		querMensagem    string
	}{
		{"documento obrigatorio", " ", "", []ItemEntrada{item}, ErrDocumentoOrigemObrigatorio, ""},
		{"fornecedor invalido", "NF-123", "invalido", []ItemEntrada{item}, ErrFornecedorIDInvalido, ""},
		{"itens obrigatorios", "NF-123", "", nil, ErrItensObrigatorios, ""},
		{"item repetido", "NF-123", "", []ItemEntrada{item, item}, ErrItemRepetido, ""},
		{"quantidade invalida", "NF-123", "", []ItemEntrada{{ItemID: item.ItemID, Quantidade: 0, CustoUnitario: 10}}, nil, "quantidade deve ser maior que zero"},
		{"custo invalido", "NF-123", "", []ItemEntrada{{ItemID: item.ItemID, Quantidade: 1, CustoUnitario: 0}}, ErrCustoInvalido, ""},
	}

	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			_, err := NovaEntradaCadastro(test.documentoOrigem, test.fornecedorID, "", false, test.itens)
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

func TestNovaEntradaCadastroLimiteDeItens(t *testing.T) {
	itens := make([]ItemEntrada, 201)
	for i := range itens {
		itens[i] = ItemEntrada{
			ItemID:        fmt.Sprintf("item-%d", i),
			Quantidade:    1,
			CustoUnitario: 1,
		}
	}

	_, err := NovaEntradaCadastro("NF-123", "", "", false, itens)
	if !errors.Is(err, ErrItensExcedemLimite) {
		t.Fatalf("erro=%v", err)
	}
}
