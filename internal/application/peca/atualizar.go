package peca

import (
	"context"
	"errors"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var ErrVersaoDivergente = errors.New("a peca foi alterada por outra requisicao; recarregue e tente de novo")

type AtualizarRepository interface {
	Atualizar(ctx context.Context, id string, version int, atualizacao peca.Atualizacao, usuarioID string) (peca.Peca, error)
}

type AtualizarPeca struct {
	repository AtualizarRepository
}

func NewAtualizarPeca(repository AtualizarRepository) AtualizarPeca {
	return AtualizarPeca{repository: repository}
}

func (useCase AtualizarPeca) Execute(ctx context.Context, id string, version int, atualizacao peca.Atualizacao, usuarioID string) (peca.Peca, error) {
	if !validation.IsUUID(id) {
		return peca.Peca{}, ErrIdentificadorInvalido
	}
	if !validation.IsUUID(atualizacao.CategoriaID) {
		return peca.Peca{}, ErrIdentificadorInvalido
	}
	return useCase.repository.Atualizar(ctx, id, version, atualizacao, usuarioID)
}
