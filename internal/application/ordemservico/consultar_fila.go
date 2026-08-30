package ordemservico

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type ConsultarFilaInput struct {
	Limite       int
	Deslocamento int
}

type ResultadoFila struct {
	Itens          []domain.ItemFila
	TotalElementos int
}

type ConsultarFilaRepository interface {
	ConsultarFila(ctx context.Context, limite, deslocamento int) ([]domain.ItemFila, int, error)
}

type ConsultarFila struct{ repository ConsultarFilaRepository }

func NewConsultarFila(repository ConsultarFilaRepository) ConsultarFila {
	return ConsultarFila{repository: repository}
}

func (useCase ConsultarFila) Execute(ctx context.Context, input ConsultarFilaInput) (ResultadoFila, error) {
	itens, total, err := useCase.repository.ConsultarFila(ctx, input.Limite, input.Deslocamento)
	if err != nil {
		return ResultadoFila{}, err
	}
	return ResultadoFila{Itens: itens, TotalElementos: total}, nil
}
