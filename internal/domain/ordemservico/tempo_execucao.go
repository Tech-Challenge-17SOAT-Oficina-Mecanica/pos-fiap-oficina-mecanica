package ordemservico

import (
	"errors"
	"time"
)

var ErrTempoExecucaoIndisponivel = errors.New("ordem de servico nao possui inicio e finalizacao da execucao")

type TempoExecucao struct {
	OrdemServicoID       string
	DataInicioExecucao   time.Time
	DataFinalizacao      time.Time
	TempoExecucaoMinutos int
}

func NovoTempoExecucao(ordemServicoID string, dataInicioExecucao, dataFinalizacao time.Time) TempoExecucao {
	return TempoExecucao{
		OrdemServicoID:       ordemServicoID,
		DataInicioExecucao:   dataInicioExecucao,
		DataFinalizacao:      dataFinalizacao,
		TempoExecucaoMinutos: int(dataFinalizacao.Sub(dataInicioExecucao).Minutes()),
	}
}
