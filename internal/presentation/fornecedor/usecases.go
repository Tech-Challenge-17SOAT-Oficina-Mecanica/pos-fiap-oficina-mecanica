package fornecedor

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type AtualizarFornecedorUseCase interface {
	Execute(ctx context.Context, fornecedorID string, atualizacao domain.Atualizacao, version int, usuarioID string) (domain.Fornecedor, error)
}

type CadastrarUseCase interface {
	Execute(ctx context.Context, cadastro domain.Cadastro) (domain.Fornecedor, error)
}

type ConsultarFornecedorPorIDUseCase interface {
	Execute(ctx context.Context, fornecedorID string) (domain.Fornecedor, error)
}

type ConsultarFornecedoresUseCase interface {
	Execute(ctx context.Context, filtros application.FiltrosConsulta) (application.ResultadoConsulta, error)
}

type DesativarFornecedorUseCase interface {
	Execute(ctx context.Context, fornecedorID, usuarioID string) (domain.Fornecedor, error)
}

type ReativarFornecedorUseCase interface {
	Execute(ctx context.Context, fornecedorID, usuarioID string) (domain.Fornecedor, error)
}
