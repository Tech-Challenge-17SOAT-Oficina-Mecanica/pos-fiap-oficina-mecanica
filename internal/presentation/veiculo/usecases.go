package veiculo

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

// Contratos de entrada consumidos pelo adaptador HTTP.
type CadastrarUseCase interface {
	Execute(context.Context, string, domain.Cadastro) (domain.Veiculo, error)
}

type ConsultarUseCase interface {
	Execute(context.Context, string, bool) (domain.Veiculo, error)
}

type AtualizarUseCase interface {
	Execute(context.Context, string, int, domain.Cadastro) (domain.Veiculo, error)
}

type InativarUseCase interface {
	Execute(context.Context, string, string, string) (application.Inativacao, error)
}

type ReativarUseCase interface {
	Execute(context.Context, string, string) (application.Reativacao, error)
}
