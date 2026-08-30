package ordemservico

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type IniciarExecucaoInput struct {
	OSID      string
	UsuarioID string
}

type IniciarExecucaoRepository interface {
	IniciarExecucao(context.Context, IniciarExecucaoInput) (domain.ResultadoInicioExecucao, error)
}

type IniciarExecucao struct{ repository IniciarExecucaoRepository }

func NewIniciarExecucao(repository IniciarExecucaoRepository) IniciarExecucao {
	return IniciarExecucao{repository: repository}
}

func (useCase IniciarExecucao) Execute(ctx context.Context, input IniciarExecucaoInput) (domain.ResultadoInicioExecucao, error) {
	return useCase.repository.IniciarExecucao(ctx, input)
}
