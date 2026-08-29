package ordemservico

import (
	"context"
	"testing"
	"time"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type tempoExecucaoRepositoryFake struct {
	itens        []domain.TempoExecucao
	total        int
	totalMinutos int64
	dataInicio   *time.Time
	dataFim      *time.Time
}

func (fake *tempoExecucaoRepositoryFake) ConsultarTempoExecucao(context.Context, string) (domain.TempoExecucao, error) {
	return domain.TempoExecucao{}, nil
}

func (fake *tempoExecucaoRepositoryFake) ListarTemposExecucao(_ context.Context, dataInicio, dataFim *time.Time, _, _ int) ([]domain.TempoExecucao, int, int64, error) {
	fake.dataInicio, fake.dataFim = dataInicio, dataFim
	return fake.itens, fake.total, fake.totalMinutos, nil
}

func TestConsultarTempoMedioExecucaoCalculaMediaDoConjuntoCompleto(t *testing.T) {
	fake := &tempoExecucaoRepositoryFake{total: 3, totalMinutos: 555}
	resultado, err := NewConsultarTempoMedioExecucao(fake).Execute(context.Background(), ConsultarTempoMedioExecucaoInput{DataInicio: "2026-08-10", DataFim: "2026-08-12"})
	if err != nil || resultado.TempoMedioExecucaoMinutos != 185 || resultado.TotalElementos != 3 {
		t.Fatalf("resultado=%+v erro=%v", resultado, err)
	}
	if fake.dataInicio == nil || fake.dataFim == nil || fake.dataInicio.Format("2006-01-02") != "2026-08-10" || fake.dataFim.Format("2006-01-02") != "2026-08-12" {
		t.Fatalf("filtros=%v,%v", fake.dataInicio, fake.dataFim)
	}
}

func TestConsultarTempoMedioExecucaoRejeitaPeriodoInvalido(t *testing.T) {
	_, err := NewConsultarTempoMedioExecucao(&tempoExecucaoRepositoryFake{}).Execute(context.Background(), ConsultarTempoMedioExecucaoInput{DataInicio: "2026-08-12", DataFim: "2026-08-10"})
	if err != ErrPeriodoInvalido {
		t.Fatalf("erro=%v", err)
	}
}
