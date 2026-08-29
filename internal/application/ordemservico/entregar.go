package ordemservico

import (
	"context"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type EntregarInput struct {
	OSID        string
	ClienteID   string
	Observacoes string
	UsuarioID   string
}

type EntregarRepository interface {
	Entregar(context.Context, EntregarInput) (domain.ResultadoEntrega, error)
}

type Entregar struct{ repository EntregarRepository }

func NewEntregar(repository EntregarRepository) Entregar { return Entregar{repository: repository} }

func (useCase Entregar) Execute(ctx context.Context, input EntregarInput) (domain.ResultadoEntrega, error) {
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	input.Observacoes = strings.TrimSpace(input.Observacoes)
	if input.ClienteID != "" && !validation.IsUUID(input.ClienteID) {
		return domain.ResultadoEntrega{}, ErrClienteIDInvalido
	}
	return useCase.repository.Entregar(ctx, input)
}
