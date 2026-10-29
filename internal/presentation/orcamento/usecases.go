package orcamento

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type AprovarUseCase interface {
	Execute(ctx context.Context, input application.AprovarInput) (domain.Aprovacao, error)
}

type CalcularUseCase interface {
	Execute(ctx context.Context, orcamentoID string) (application.ResultadoCalculo, error)
}

type ConsultarUseCase interface {
	Execute(ctx context.Context, ordemServicoID, clienteID string) (domain.Consulta, error)
}

type EnviarUseCase interface {
	Execute(ctx context.Context, orcamentoID, usuarioID string) (application.ResultadoEnvio, error)
}

type RecusarUseCase interface {
	Execute(ctx context.Context, input application.RecusarInput) (domain.Decisao, error)
}
