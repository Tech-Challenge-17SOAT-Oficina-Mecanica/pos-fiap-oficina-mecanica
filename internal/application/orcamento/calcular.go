package orcamento

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrOrcamentoNaoEncontrado = errors.New("orçamento não encontrado")
	ErrFalhaPersistencia      = errors.New("falha ao persistir cálculo do orçamento")
)

type Repository interface {
	Calcular(context.Context, string, string) (json.Number, error)
}

type Calcular struct{ repository Repository }

func NewCalcular(repository Repository) Calcular { return Calcular{repository: repository} }

func (useCase Calcular) Execute(ctx context.Context, orcamentoID, usuarioID string) (json.Number, error) {
	return useCase.repository.Calcular(ctx, orcamentoID, usuarioID)
}
