package cliente

import (
	"context"
	"errors"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

var ErrClienteDuplicado = errors.New("cliente já cadastrado com o CPF/CNPJ informado")

type Repository interface {
	ExisteAtivoPorDocumento(context.Context, string) (bool, error)
	Salvar(context.Context, cliente.Cliente) (cliente.Cliente, error)
}

type Cadastrar struct {
	repository Repository
}

func NewCadastrar(repository Repository) Cadastrar {
	return Cadastrar{repository: repository}
}

func (useCase Cadastrar) Execute(ctx context.Context, input cliente.NovoClienteInput) (cliente.Cliente, error) {
	novo, err := cliente.Novo(input)
	if err != nil {
		return cliente.Cliente{}, err
	}
	exists, err := useCase.repository.ExisteAtivoPorDocumento(ctx, novo.Documento)
	if err != nil {
		return cliente.Cliente{}, err
	}
	if exists {
		return cliente.Cliente{}, ErrClienteDuplicado
	}
	return useCase.repository.Salvar(ctx, novo)
}
