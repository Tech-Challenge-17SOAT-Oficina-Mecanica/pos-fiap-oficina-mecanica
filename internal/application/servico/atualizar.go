package servico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

var ErrVersaoDivergente = errors.New("If-Match divergente")

type AtualizarRepository interface {
	BuscarPorID(context.Context, string) (domain.Servico, error)
	ExisteAtivoPorNomeNormalizadoExcetoID(context.Context, string, string) (bool, error)
	Atualizar(context.Context, domain.Servico, int, string) (domain.Servico, error)
}

type Atualizar struct{ repository AtualizarRepository }

func NewAtualizar(repository AtualizarRepository) Atualizar { return Atualizar{repository: repository} }

func (useCase Atualizar) Execute(ctx context.Context, id string, version int, atualizacao domain.Atualizacao, usuarioID string) (domain.Servico, error) {
	atual, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return domain.Servico{}, err
	}
	if atual.Version != version {
		return domain.Servico{}, ErrVersaoDivergente
	}
	resultado, err := atual.Atualizar(atualizacao)
	if err != nil {
		return domain.Servico{}, err
	}
	existe, err := useCase.repository.ExisteAtivoPorNomeNormalizadoExcetoID(ctx, resultado.NomeNormalizado, id)
	if err != nil {
		return domain.Servico{}, err
	}
	if existe {
		return domain.Servico{}, ErrServicoDuplicado
	}
	return useCase.repository.Atualizar(ctx, resultado, version, usuarioID)
}
