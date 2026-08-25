package veiculo

import (
	"context"
	"errors"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

var ErrVersaoDivergente = errors.New("versão do veículo divergente")

type Atualizar struct{ repository Repository }

func NewAtualizar(repository Repository) Atualizar { return Atualizar{repository} }
func (useCase Atualizar) Execute(ctx context.Context, id string, version int, cadastro domain.Cadastro) (domain.Veiculo, error) {
	return useCase.repository.Atualizar(ctx, id, version, cadastro)
}
