package veiculo

import (
	"context"
	"errors"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

var (
	ErrClienteNaoEncontrado = errors.New("cliente não encontrado")
	ErrClienteInativo       = errors.New("cliente inativo")
	ErrPlacaDuplicada       = errors.New("placa já cadastrada")
)

type Repository interface {
	CadastrarParaCliente(context.Context, string, domain.Cadastro) (domain.Veiculo, error)
}
type Cadastrar struct{ repository Repository }

func NewCadastrar(repository Repository) Cadastrar { return Cadastrar{repository} }
func (useCase Cadastrar) Execute(ctx context.Context, clienteID string, cadastro domain.Cadastro) (domain.Veiculo, error) {
	return useCase.repository.CadastrarParaCliente(ctx, clienteID, cadastro)
}
