package orcamento

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrIdentificadorInvalido       = errors.New("orcamentoId deve ser um UUID valido")
	ErrOrcamentoJaDecidido         = errors.New("orcamento ja aprovado ou recusado")
	ErrOrdemServicoStatusInvalido  = errors.New("ordem de servico fora de AGUARDANDO_APROVACAO")
	ErrOrcamentoComplementarSemPai = errors.New("orcamento complementar sem vinculo valido com o principal")
)

type AprovarInput struct {
	OrcamentoID    string
	ClienteID      string
	UsuarioID      string
	OrdemServicoID string
}

type AprovarRepository interface {
	Aprovar(context.Context, AprovarInput) (domain.Aprovacao, error)
}

type Aprovar struct{ repository AprovarRepository }

func NewAprovar(repository AprovarRepository) Aprovar {
	return Aprovar{repository: repository}
}

func (useCase Aprovar) Execute(ctx context.Context, input AprovarInput) (domain.Aprovacao, error) {
	input.OrcamentoID = strings.TrimSpace(input.OrcamentoID)
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	input.UsuarioID = strings.TrimSpace(input.UsuarioID)
	input.OrdemServicoID = strings.TrimSpace(input.OrdemServicoID)

	if !validation.IsUUID(input.OrcamentoID) {
		return domain.Aprovacao{}, ErrIdentificadorInvalido
	}
	if input.ClienteID != "" && !validation.IsUUID(input.ClienteID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	if input.ClienteID == "" && !validation.IsUUID(input.UsuarioID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	if input.OrdemServicoID != "" && !validation.IsUUID(input.OrdemServicoID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	return useCase.repository.Aprovar(ctx, input)
}
