package insumo

import (
	"context"
	"errors"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrIdentificadorInvalido = errors.New("identificador deve ser um UUID valido")
	ErrCategoriaInvalida     = errors.New("categoria inexistente ou inativa")
	ErrDescricaoDuplicada    = errors.New("ja existe insumo ativo com esta descricao na categoria e unidade")
)

type CadastrarRepository interface {
	Cadastrar(ctx context.Context, cadastro insumo.Cadastro) (insumo.Insumo, error)
}

type CadastrarInsumo struct {
	repository CadastrarRepository
}

func NewCadastrarInsumo(repository CadastrarRepository) CadastrarInsumo {
	return CadastrarInsumo{repository: repository}
}

func (useCase CadastrarInsumo) Execute(ctx context.Context, cadastro insumo.Cadastro) (insumo.Insumo, error) {
	if !validation.IsUUID(cadastro.CategoriaID) {
		return insumo.Insumo{}, ErrIdentificadorInvalido
	}
	return useCase.repository.Cadastrar(ctx, cadastro)
}
