package peca

import "testing"

func TestSaldoDisponivelEDisponibilidade(t *testing.T) {
	casos := []struct {
		nome            string
		saldoFisico     int64
		saldoReservado  int64
		estoqueMinimo   int64
		saldoDisponivel int64
		disponivel      bool
		abaixoDoMinimo  bool
	}{
		{"sem reserva", 10, 0, 5, 10, true, false},
		{"reserva parcial", 10, 6, 5, 4, true, true},
		{"reserva total", 10, 10, 5, 0, false, true},
		{"sem saldo e sem minimo", 0, 0, 0, 0, false, false},
		{"exatamente no minimo", 5, 0, 5, 5, true, false},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			item := Peca{SaldoFisico: caso.saldoFisico, SaldoReservado: caso.saldoReservado, EstoqueMinimo: caso.estoqueMinimo}
			if item.SaldoDisponivel() != caso.saldoDisponivel {
				t.Fatalf("SaldoDisponivel() = %d, esperado %d", item.SaldoDisponivel(), caso.saldoDisponivel)
			}
			if item.Disponivel() != caso.disponivel {
				t.Fatalf("Disponivel() = %v, esperado %v", item.Disponivel(), caso.disponivel)
			}
			if item.AbaixoDoMinimo() != caso.abaixoDoMinimo {
				t.Fatalf("AbaixoDoMinimo() = %v, esperado %v", item.AbaixoDoMinimo(), caso.abaixoDoMinimo)
			}
		})
	}
}

func TestAtendeQuantidade(t *testing.T) {
	item := Peca{SaldoFisico: 6, SaldoReservado: 2}

	casos := []struct {
		quantidade int64
		esperado   bool
	}{
		{3, true},
		{4, true},
		{5, false},
	}

	for _, caso := range casos {
		if item.AtendeQuantidade(caso.quantidade) != caso.esperado {
			t.Fatalf("AtendeQuantidade(%d) = %v, esperado %v", caso.quantidade, item.AtendeQuantidade(caso.quantidade), caso.esperado)
		}
	}
}
