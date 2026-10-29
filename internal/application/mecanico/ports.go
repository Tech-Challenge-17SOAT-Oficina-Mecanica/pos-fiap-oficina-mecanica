package mecanico

import (
	"context"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
)

type CadastrarRepository interface {
	EmailExiste(context.Context, string) (bool, error)
	SalvarMecanico(context.Context, domain.Mecanico, string) (domain.Mecanico, error)
}

type AtualizarRepository interface {
	BuscarPorID(context.Context, string) (domain.Mecanico, error)
	EmailExisteExcetoMecanico(context.Context, string, string) (bool, error)
	AtualizarMecanico(context.Context, domain.Mecanico, int) (domain.Mecanico, error)
}
