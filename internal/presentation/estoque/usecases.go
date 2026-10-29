package estoque

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
)

type RegistrarEntradaUseCase interface {
	Execute(ctx context.Context, input application.RegistrarEntradaInput) (application.Resultado, error)
}
