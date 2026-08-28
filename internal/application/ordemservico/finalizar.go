package ordemservico

import (
	"context"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type FinalizarInput struct {
	OSID        string
	Observacoes string
	UsuarioID   string
}

type FinalizarRepository interface {
	Finalizar(context.Context, FinalizarInput) (domain.ResultadoFinalizacao, error)
}

type Finalizar struct{ repository FinalizarRepository }

func NewFinalizar(repository FinalizarRepository) Finalizar {
	return Finalizar{repository: repository}
}

func (useCase Finalizar) Execute(ctx context.Context, input FinalizarInput) (domain.ResultadoFinalizacao, error) {
	input.Observacoes = strings.TrimSpace(input.Observacoes)
	return useCase.repository.Finalizar(ctx, input)
}
