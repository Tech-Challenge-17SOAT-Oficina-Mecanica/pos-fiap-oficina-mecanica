package insumo

import "testing"

func TestSaldosDoInsumo(t *testing.T) {
	item := Insumo{SaldoFisico: "10.500", SaldoReservado: "2.250", EstoqueMinimo: "8.000"}

	if got := item.SaldoDisponivel(); got != "8.25" {
		t.Fatalf("saldoDisponivel = %q", got)
	}
	if !item.Disponivel() {
		t.Fatal("esperava insumo disponível")
	}
	if item.AbaixoDoMinimo() {
		t.Fatal("não deveria estar abaixo do mínimo")
	}
	if !item.AtendeQuantidade("8.250") || item.AtendeQuantidade("8.251") {
		t.Fatalf("comparação de quantidade incorreta para saldo %s", item.SaldoDisponivel())
	}
}

func TestInsumoAbaixoDoMinimoESemDisponibilidade(t *testing.T) {
	item := Insumo{SaldoFisico: "2.000", SaldoReservado: "2.000", EstoqueMinimo: "1.000"}

	if item.SaldoDisponivel() != "0" || item.Disponivel() || !item.AbaixoDoMinimo() {
		t.Fatalf("estado de saldo incorreto: disponível=%s", item.SaldoDisponivel())
	}
}
