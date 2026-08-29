package ordemservico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var ErrAcessoNegado = errors.New("cliente sem acesso a esta ordem de servico")

type ConsultarRepository interface {
	Consultar(ctx context.Context, ordemServicoID, clienteID string) (domain.ConsultaDetalhada, error)
}

type Consultar struct{ repository ConsultarRepository }

func NewConsultar(repository ConsultarRepository) Consultar { return Consultar{repository: repository} }

func (useCase Consultar) Execute(ctx context.Context, ordemServicoID, clienteID string) (domain.ConsultaDetalhada, error) {
	return useCase.repository.Consultar(ctx, ordemServicoID, clienteID)
}
