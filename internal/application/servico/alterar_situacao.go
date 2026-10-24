package servico

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type Desativar struct{ repository DesativarRepository }

func NewDesativar(repository DesativarRepository) Desativar { return Desativar{repository: repository} }

func (useCase Desativar) Execute(ctx context.Context, id, usuarioID string) (domain.Servico, error) {
	atual, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return domain.Servico{}, err
	}
	if _, err := atual.Desativar(); err != nil {
		return domain.Servico{}, err
	}
	return useCase.repository.Desativar(ctx, id, usuarioID)
}

type Reativar struct{ repository ReativarRepository }

func NewReativar(repository ReativarRepository) Reativar { return Reativar{repository: repository} }

func (useCase Reativar) Execute(ctx context.Context, id string) (domain.Servico, error) {
	atual, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return domain.Servico{}, err
	}
	if _, err := atual.Reativar(); err != nil {
		return domain.Servico{}, err
	}
	existe, err := useCase.repository.ExisteAtivoPorNomeNormalizadoExcetoID(ctx, atual.NomeNormalizado, id)
	if err != nil {
		return domain.Servico{}, err
	}
	if existe {
		return domain.Servico{}, ErrServicoDuplicado
	}
	return useCase.repository.Reativar(ctx, id)
}
