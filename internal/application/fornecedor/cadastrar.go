package fornecedor

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

var ErrDocumentoDuplicado = errors.New("ja existe fornecedor ativo com este documento")

type Repository interface {
	Cadastrar(context.Context, domain.Cadastro) (domain.Fornecedor, error)
}

type Cadastrar struct {
	repository Repository
}

func NewCadastrar(repository Repository) Cadastrar {
	return Cadastrar{repository: repository}
}

func (useCase Cadastrar) Execute(ctx context.Context, cadastro domain.Cadastro) (domain.Fornecedor, error) {
	return useCase.repository.Cadastrar(ctx, cadastro)
}
