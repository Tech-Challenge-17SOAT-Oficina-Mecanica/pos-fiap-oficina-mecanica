package orcamento

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

var (
	ErrOrcamentoJaDecidido               = errors.New("orcamento ja foi aprovado ou recusado")
	ErrOrcamentoComplementarSemPrincipal = errors.New("orcamento complementar sem orcamento principal vinculado")
	ErrMotivoInvalido                    = errors.New("motivo deve ter no maximo 500 caracteres")
)

type RecusarInput struct {
	OrcamentoID    string
	ClienteID      string
	OrdemServicoID string
	Motivo         string
}

type RecusarRepository interface {
	Recusar(context.Context, RecusarInput) (domain.Decisao, error)
}

type Recusar struct{ repository RecusarRepository }

func NewRecusar(repository RecusarRepository) Recusar { return Recusar{repository: repository} }

func (useCase Recusar) Execute(ctx context.Context, input RecusarInput) (domain.Decisao, error) {
	input.Motivo = strings.TrimSpace(input.Motivo)
	if len(input.Motivo) > 500 {
		return domain.Decisao{}, ErrMotivoInvalido
	}
	return useCase.repository.Recusar(ctx, input)
}
