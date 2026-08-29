package insumo

import "testing"

func TestNovoProcessamentoInsumo(t *testing.T) {
	casos := []struct {
		nome       string
		quantidade string
		saldo      string
		reserva    string
		compra     string
		apos       string
	}{
		{"total", "2.5", "10", "2.5", "0", "7.5"},
		{"parcial", "4", "1.25", "1.25", "2.75", "0"},
		{"sem saldo", "3", "0", "0", "3", "0"},
		{"saldo negativo tratado como zero", "3", "-1", "0", "3", "-1"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			processamento := NovoProcessamento("item", caso.quantidade, caso.saldo)
			if processamento.QuantidadeReservada != caso.reserva ||
				processamento.QuantidadeCompra != caso.compra ||
				processamento.SaldoDisponivelApos != caso.apos {
				t.Fatalf("processamento=%+v", processamento)
			}
		})
	}
}
