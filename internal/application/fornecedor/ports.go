package fornecedor

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type ListarRepository interface {
	Listar(context.Context, FiltrosConsulta) ([]domain.Fornecedor, int, error)
}

type BuscarPorIDRepository interface {
	BuscarPorID(context.Context, string) (domain.Fornecedor, error)
}

type DesativarRepository interface {
	Desativar(context.Context, string, string) (domain.Fornecedor, error)
}

type ReativarRepository interface {
	Reativar(context.Context, string, string) (domain.Fornecedor, error)
}
