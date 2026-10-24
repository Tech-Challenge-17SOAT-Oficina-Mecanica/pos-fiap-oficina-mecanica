package ordemservico

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	orcamentoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type ConsultarFilaUseCase interface {
	Execute(ctx context.Context, input application.ConsultarFilaInput) (application.ResultadoFila, error)
}

type ConsultarTempoExecucaoDaOSUseCase interface {
	Execute(ctx context.Context, ordemServicoID string) (domain.TempoExecucao, error)
}

type ConsultarTempoMedioExecucaoUseCase interface {
	Execute(ctx context.Context, input application.ConsultarTempoMedioExecucaoInput) (application.ResultadoTempoMedioExecucao, error)
}

type ConsultarUseCase interface {
	Execute(ctx context.Context, ordemServicoID, clienteID string) (domain.ConsultaDetalhada, error)
}

type EntregarUseCase interface {
	Execute(ctx context.Context, input application.EntregarInput) (domain.ResultadoEntrega, error)
}

type FinalizarUseCase interface {
	Execute(ctx context.Context, input application.FinalizarInput) (domain.ResultadoFinalizacao, error)
}

type IniciarExecucaoUseCase interface {
	Execute(ctx context.Context, input application.IniciarExecucaoInput) (domain.ResultadoInicioExecucao, error)
}

type ListarUseCase interface {
	Execute(ctx context.Context, input application.ListarInput) (application.ResultadoListagem, error)
}

type RegistrarItensUseCase interface {
	Execute(ctx context.Context, input application.RegistrarInput) (orcamentoDomain.Resultado, error)
}

type RegistrarProblemaRelatadoUseCase interface {
	Execute(ctx context.Context, input application.RegistrarProblemaRelatadoInput) (domain.OrdemDeServico, error)
}

type RegistrarProblemaUseCase interface {
	Execute(ctx context.Context, ordemServicoID string, cadastro domain.ProblemaCadastro) (application.ResultadoRegistroProblema, error)
}

type RegistrarServicosUseCase interface {
	Execute(ctx context.Context, ordemServicoID string, servicos []domain.ServicoCadastro) (application.ResultadoRegistroServicos, error)
}
