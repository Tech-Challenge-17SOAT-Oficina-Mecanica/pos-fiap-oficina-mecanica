package ordemservico

import (
	"context"
	"errors"
	"time"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var (
	ErrDataInicioInvalida = errors.New("dataInicio invalida")
	ErrDataFimInvalida    = errors.New("dataFim invalida")
	ErrPeriodoInvalido    = errors.New("dataInicio nao pode ser posterior a dataFim")
)

type ConsultarTempoExecucaoRepository interface {
	ConsultarTempoExecucao(context.Context, string) (domain.TempoExecucao, error)
	ListarTemposExecucao(context.Context, *time.Time, *time.Time, int, int) ([]domain.TempoExecucao, int, int64, error)
}

type ConsultarTempoExecucaoDaOS struct {
	repository ConsultarTempoExecucaoRepository
}

func NewConsultarTempoExecucaoDaOS(repository ConsultarTempoExecucaoRepository) ConsultarTempoExecucaoDaOS {
	return ConsultarTempoExecucaoDaOS{repository: repository}
}

func (useCase ConsultarTempoExecucaoDaOS) Execute(ctx context.Context, ordemServicoID string) (domain.TempoExecucao, error) {
	return useCase.repository.ConsultarTempoExecucao(ctx, ordemServicoID)
}

type ConsultarTempoMedioExecucao struct {
	repository ConsultarTempoExecucaoRepository
}

type ConsultarTempoMedioExecucaoInput struct {
	DataInicio   string
	DataFim      string
	Limite       int
	Deslocamento int
}

type ResultadoTempoMedioExecucao struct {
	Itens                     []domain.TempoExecucao
	TotalElementos            int
	TempoMedioExecucaoMinutos int
}

func NewConsultarTempoMedioExecucao(repository ConsultarTempoExecucaoRepository) ConsultarTempoMedioExecucao {
	return ConsultarTempoMedioExecucao{repository: repository}
}

func (useCase ConsultarTempoMedioExecucao) Execute(ctx context.Context, input ConsultarTempoMedioExecucaoInput) (ResultadoTempoMedioExecucao, error) {
	dataInicio, err := dataFiltro(input.DataInicio, ErrDataInicioInvalida)
	if err != nil {
		return ResultadoTempoMedioExecucao{}, err
	}
	dataFim, err := dataFiltro(input.DataFim, ErrDataFimInvalida)
	if err != nil {
		return ResultadoTempoMedioExecucao{}, err
	}
	if dataInicio != nil && dataFim != nil && dataInicio.After(*dataFim) {
		return ResultadoTempoMedioExecucao{}, ErrPeriodoInvalido
	}
	itens, total, totalMinutos, err := useCase.repository.ListarTemposExecucao(ctx, dataInicio, dataFim, input.Limite, input.Deslocamento)
	if err != nil {
		return ResultadoTempoMedioExecucao{}, err
	}
	resultado := ResultadoTempoMedioExecucao{Itens: itens, TotalElementos: total}
	if total > 0 {
		resultado.TempoMedioExecucaoMinutos = int(totalMinutos / int64(total))
	}
	return resultado, nil
}

func dataFiltro(valor string, erro error) (*time.Time, error) {
	if valor == "" {
		return nil, nil
	}
	data, parseErr := time.Parse("2006-01-02", valor)
	if parseErr != nil {
		return nil, erro
	}
	return &data, nil
}
