package cliente

import (
	"context"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type AtualizarInput struct {
	ClienteID string
	Version   int
	Dados     cliente.AtualizarClienteInput
}

type Atualizar struct {
	repository Repository
}

func NewAtualizar(repository Repository) Atualizar {
	return Atualizar{repository: repository}
}

func (useCase Atualizar) Execute(ctx context.Context, input AtualizarInput) (cliente.Cliente, error) {
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	if input.ClienteID == "" {
		return cliente.Cliente{}, cliente.ErrClienteIDObrigatorio
	}
	atual, err := useCase.repository.BuscarPorID(ctx, input.ClienteID)
	if err != nil {
		return cliente.Cliente{}, err
	}
	if atual.Version != input.Version {
		return cliente.Cliente{}, ErrVersaoDivergente
	}
	atualizado, err := atual.Atualizar(input.Dados)
	if err != nil {
		return cliente.Cliente{}, err
	}
	exists, err := useCase.repository.ExisteAtivoPorDocumentoExcetoID(ctx, atualizado.Documento, atualizado.ID)
	if err != nil {
		return cliente.Cliente{}, err
	}
	if exists {
		return cliente.Cliente{}, ErrClienteDuplicado
	}
	return useCase.repository.Atualizar(ctx, atualizado, input.Version)
}
