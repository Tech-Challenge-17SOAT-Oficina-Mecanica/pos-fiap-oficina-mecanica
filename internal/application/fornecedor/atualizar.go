package fornecedor

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

var (
	ErrAtualizacaoInvalida = errors.New("dados de atualizacao invalidos")
	ErrFornecedorInativo   = errors.New("fornecedor inativo")
	ErrVersaoDivergente    = errors.New("fornecedor alterado por outro usuario")
)

type AtualizacaoRepository interface {
	Atualizar(context.Context, string, domain.Atualizacao, int, string) (domain.Fornecedor, error)
}

type AtualizarFornecedor struct {
	repository AtualizacaoRepository
}

func NewAtualizarFornecedor(repository AtualizacaoRepository) AtualizarFornecedor {
	return AtualizarFornecedor{repository: repository}
}

func (useCase AtualizarFornecedor) Execute(ctx context.Context, fornecedorID string, atualizacao domain.Atualizacao, version int, usuarioID string) (domain.Fornecedor, error) {
	fornecedorID = strings.TrimSpace(fornecedorID)
	usuarioID = strings.TrimSpace(usuarioID)
	if fornecedorID == "" || version < 1 {
		return domain.Fornecedor{}, ErrAtualizacaoInvalida
	}
	fornecedor, err := useCase.repository.Atualizar(ctx, fornecedorID, atualizacao, version, usuarioID)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}
