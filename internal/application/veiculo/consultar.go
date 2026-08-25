package veiculo

import (
	"context"
	"errors"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

var ErrVeiculoNaoEncontrado = errors.New("veículo não encontrado")

type Consultar struct{ repository Repository }

func NewConsultar(repository Repository) Consultar { return Consultar{repository} }
func (useCase Consultar) Execute(ctx context.Context, placa string, incluirInativos bool) (domain.Veiculo, error) {
	return useCase.repository.ConsultarPorPlaca(ctx, placa, incluirInativos)
}
