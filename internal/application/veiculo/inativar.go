package veiculo

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

type Inativar struct{ repository Repository }

func NewInativar(repository Repository) Inativar { return Inativar{repository} }

func (useCase Inativar) Execute(ctx context.Context, veiculoID, usuarioID, motivo string) (Inativacao, error) {
	motivo, err := domain.MotivoParaInativacao(motivo)
	if err != nil {
		return Inativacao{}, err
	}
	veiculo, err := useCase.repository.BuscarPorIDIncluindoInativo(ctx, veiculoID)
	if err != nil {
		return Inativacao{}, err
	}
	if !veiculo.Ativo {
		return Inativacao{}, ErrVeiculoJaInativo
	}
	ordens, err := useCase.repository.BuscarOSAbertas(ctx, veiculoID)
	if err != nil {
		return Inativacao{}, err
	}
	if len(ordens) > 0 {
		return Inativacao{}, OSAbertaError{Ordens: ordens}
	}
	return useCase.repository.Inativar(ctx, InativarRepositoryInput{VeiculoID: veiculoID, InativadoPor: usuarioID, Motivo: motivo})
}

type Reativar struct{ repository Repository }

func NewReativar(repository Repository) Reativar { return Reativar{repository} }

func (useCase Reativar) Execute(ctx context.Context, veiculoID, usuarioID string) (Reativacao, error) {
	veiculo, err := useCase.repository.BuscarPorIDIncluindoInativo(ctx, veiculoID)
	if err != nil {
		return Reativacao{}, err
	}
	if veiculo.Ativo {
		return Reativacao{}, ErrVeiculoJaAtivo
	}
	clienteAtivo, err := useCase.repository.ClienteAtivo(ctx, veiculo.ClienteID)
	if err != nil {
		return Reativacao{}, err
	}
	if !clienteAtivo {
		return Reativacao{}, ErrClienteProprietarioInativo
	}
	existe, err := useCase.repository.ExisteAtivoPorPlacaExcetoID(ctx, veiculo.Placa, veiculo.ID)
	if err != nil {
		return Reativacao{}, err
	}
	if existe {
		return Reativacao{}, ErrPlacaDuplicada
	}
	return useCase.repository.Reativar(ctx, veiculoID, usuarioID)
}
