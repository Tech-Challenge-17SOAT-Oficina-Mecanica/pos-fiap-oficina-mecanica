package ordemservico

import (
	"context"
	"time"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type ConsultarTempoExecucaoRepository interface {
	ConsultarTempoExecucao(context.Context, string) (domain.TempoExecucao, error)
}

type ListarTemposExecucaoRepository interface {
	ListarTemposExecucao(context.Context, *time.Time, *time.Time, int, int) ([]domain.TempoExecucao, int, int64, error)
}
