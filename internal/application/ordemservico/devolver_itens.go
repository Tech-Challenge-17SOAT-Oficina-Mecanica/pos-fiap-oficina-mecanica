package ordemservico

import (
	"context"

	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

type DevolucaoRepository interface {
	DevolverItensAoEstoque(ctx context.Context, ordemServicoID string) (domainEstoque.ResultadoDevolucao, error)
}

// DevolverItensAoEstoque libera reservas e retorna ao estoque as pecas e insumos de uma OS.
// Sem endpoint proprio: e chamado em processo pelo caso de uso que cancela a OS (ex.: recusa de orcamento).
type DevolverItensAoEstoque struct{ repository DevolucaoRepository }

func NewDevolverItensAoEstoque(repository DevolucaoRepository) DevolverItensAoEstoque {
	return DevolverItensAoEstoque{repository: repository}
}

func (useCase DevolverItensAoEstoque) Execute(ctx context.Context, ordemServicoID string) (domainEstoque.ResultadoDevolucao, error) {
	return useCase.repository.DevolverItensAoEstoque(ctx, ordemServicoID)
}
