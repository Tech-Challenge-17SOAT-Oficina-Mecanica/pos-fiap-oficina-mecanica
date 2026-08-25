package fornecedor

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

var (
	ErrFornecedorJaInativo       = errors.New("fornecedor ja inativo")
	ErrFornecedorJaAtivo         = errors.New("fornecedor ja ativo")
	ErrFornecedorComPedidoAberto = errors.New("fornecedor possui pedido de compra em aberto")
	ErrDocumentoAtivoDuplicado   = errors.New("ja existe fornecedor ativo com este documento")
	ErrSituacaoInvalida          = errors.New("situacao do fornecedor invalida")
)

type SituacaoRepository interface {
	Desativar(context.Context, string, string) (domain.Fornecedor, error)
	Reativar(context.Context, string, string) (domain.Fornecedor, error)
}

type DesativarFornecedor struct {
	repository SituacaoRepository
}

func NewDesativarFornecedor(repository SituacaoRepository) DesativarFornecedor {
	return DesativarFornecedor{repository: repository}
}

func (useCase DesativarFornecedor) Execute(ctx context.Context, fornecedorID, usuarioID string) (domain.Fornecedor, error) {
	fornecedorID = strings.TrimSpace(fornecedorID)
	usuarioID = strings.TrimSpace(usuarioID)
	if fornecedorID == "" {
		return domain.Fornecedor{}, ErrSituacaoInvalida
	}
	return useCase.repository.Desativar(ctx, fornecedorID, usuarioID)
}

type ReativarFornecedor struct {
	repository SituacaoRepository
}

func NewReativarFornecedor(repository SituacaoRepository) ReativarFornecedor {
	return ReativarFornecedor{repository: repository}
}

func (useCase ReativarFornecedor) Execute(ctx context.Context, fornecedorID, usuarioID string) (domain.Fornecedor, error) {
	fornecedorID = strings.TrimSpace(fornecedorID)
	usuarioID = strings.TrimSpace(usuarioID)
	if fornecedorID == "" {
		return domain.Fornecedor{}, ErrSituacaoInvalida
	}
	return useCase.repository.Reativar(ctx, fornecedorID, usuarioID)
}
