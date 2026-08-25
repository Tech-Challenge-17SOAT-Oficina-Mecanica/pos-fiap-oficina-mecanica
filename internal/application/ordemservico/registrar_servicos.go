package ordemservico

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type ResultadoRegistroServicos struct {
	Orcamento domain.Orcamento
	Servicos  []domain.ServicoRegistrado
}

type ServicoRepository interface {
	RegistrarServicos(context.Context, string, []domain.ServicoCadastro) (ResultadoRegistroServicos, error)
}

type RegistrarServicos struct{ repository ServicoRepository }

func NewRegistrarServicos(repository ServicoRepository) RegistrarServicos {
	return RegistrarServicos{repository: repository}
}

func (useCase RegistrarServicos) Execute(ctx context.Context, ordemServicoID string, servicos []domain.ServicoCadastro) (ResultadoRegistroServicos, error) {
	return useCase.repository.RegistrarServicos(ctx, ordemServicoID, servicos)
}
