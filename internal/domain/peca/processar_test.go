package peca

import "testing"

func TestNovoProcessamento(t *testing.T) {
	casos := []struct {
		nome                         string
		quantidade, saldo            int64
		reservada, compra, saldoApos int64
	}{
		{"saldo total", 2, 5, 2, 0, 3},
		{"saldo parcial", 5, 2, 2, 3, 0},
		{"sem saldo", 3, 0, 0, 3, 0},
		{"saldo negativo defensivo", 3, -1, 0, 3, -1},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			resultado := NovoProcessamento("item", caso.quantidade, caso.saldo)
			if resultado.QuantidadeReservada != caso.reservada ||
				resultado.QuantidadeCompra != caso.compra ||
				resultado.SaldoDisponivelApos != caso.saldoApos {
				t.Fatalf("resultado = %+v", resultado)
			}
		})
	}
}
