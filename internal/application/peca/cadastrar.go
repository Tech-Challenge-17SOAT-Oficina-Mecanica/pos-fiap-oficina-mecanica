package peca

import (
	"context"
	"errors"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrCategoriaInvalida  = errors.New("categoria inexistente ou inativa")
	ErrFornecedorInvalido = errors.New("fornecedor inexistente ou inativo")
	ErrDescricaoDuplicada = errors.New("ja existe peca ativa com esta descricao na categoria")
)

type CadastrarRepository interface {
	Cadastrar(ctx context.Context, cadastro peca.Cadastro) (peca.Peca, error)
}

type CadastrarPeca struct {
	repository CadastrarRepository
}

func NewCadastrarPeca(repository CadastrarRepository) CadastrarPeca {
	return CadastrarPeca{repository: repository}
}

func (useCase CadastrarPeca) Execute(ctx context.Context, cadastro peca.Cadastro) (peca.Peca, error) {
	if !validation.IsUUID(cadastro.CategoriaID) {
		return peca.Peca{}, ErrIdentificadorInvalido
	}
	if cadastro.FornecedorID != nil && !validation.IsUUID(*cadastro.FornecedorID) {
		return peca.Peca{}, ErrFornecedorInvalido
	}
	return useCase.repository.Cadastrar(ctx, cadastro)
}
