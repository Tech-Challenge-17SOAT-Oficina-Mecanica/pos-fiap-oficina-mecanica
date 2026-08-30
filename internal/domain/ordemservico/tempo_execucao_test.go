package ordemservico

import (
	"errors"
	"testing"
	"time"
)

func TestNovoTempoExecucaoCalculaDiferencaEmMinutos(t *testing.T) {
	inicio := time.Date(2026, time.August, 10, 9, 0, 0, 0, time.UTC)
	fim := inicio.Add(210 * time.Minute)
	resultado, err := NovoTempoExecucao("os-1", inicio, fim)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.TempoExecucaoMinutos != 210 || resultado.OrdemServicoID != "os-1" {
		t.Fatalf("resultado=%+v", resultado)
	}
}

func TestNovoTempoExecucaoRejeitaFinalizacaoAntesDoInicio(t *testing.T) {
	inicio := time.Date(2026, time.August, 10, 9, 0, 0, 0, time.UTC)
	_, err := NovoTempoExecucao("os-1", inicio, inicio.Add(-time.Minute))
	if !errors.Is(err, ErrTempoExecucaoInvalido) {
		t.Fatalf("erro=%v, esperado %v", err, ErrTempoExecucaoInvalido)
	}
}
