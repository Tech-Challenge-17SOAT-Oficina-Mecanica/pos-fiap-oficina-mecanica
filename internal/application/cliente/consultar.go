package cliente

import (
	"context"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type Consultar struct {
	repository Repository
}

func NewConsultar(repository Repository) Consultar {
	return Consultar{repository: repository}
}

func (useCase Consultar) Execute(ctx context.Context, documento string) (cliente.Cliente, error) {
	documento, err := cliente.DocumentoParaConsulta(documento)
	if err != nil {
		return cliente.Cliente{}, err
	}
	return useCase.repository.BuscarPorDocumento(ctx, documento)
}
