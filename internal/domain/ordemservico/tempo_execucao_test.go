package ordemservico

import (
	"testing"
	"time"
)

func TestNovoTempoExecucaoCalculaDiferencaEmMinutos(t *testing.T) {
	inicio := time.Date(2026, time.August, 10, 9, 0, 0, 0, time.UTC)
	fim := inicio.Add(210 * time.Minute)
	resultado := NovoTempoExecucao("os-1", inicio, fim)
	if resultado.TempoExecucaoMinutos != 210 || resultado.OrdemServicoID != "os-1" {
		t.Fatalf("resultado=%+v", resultado)
	}
}
