package servico

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type DesativarRepository interface {
	BuscarPorID(context.Context, string) (domain.Servico, error)
	Desativar(context.Context, string, string) (domain.Servico, error)
}

type ReativarRepository interface {
	BuscarPorID(context.Context, string) (domain.Servico, error)
	ExisteAtivoPorNomeNormalizadoExcetoID(context.Context, string, string) (bool, error)
	Reativar(context.Context, string) (domain.Servico, error)
}
