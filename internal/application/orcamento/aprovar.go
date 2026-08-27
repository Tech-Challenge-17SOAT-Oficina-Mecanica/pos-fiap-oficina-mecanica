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
	ErrFornecedorIDInvalido        = errors.New("fornecedorId deve ser um UUID valido")
	ErrFornecedorNaoEncontrado     = errors.New("fornecedor inexistente")
	ErrFornecedorInativo           = errors.New("fornecedor inativo")
	ErrOrcamentoJaDecidido         = errors.New("orcamento ja aprovado ou recusado")
	ErrOrdemServicoStatusInvalido  = errors.New("ordem de servico fora de AGUARDANDO_APROVACAO")
	ErrOrcamentoComplementarSemPai = errors.New("orcamento complementar sem vinculo valido com o principal")
)

type AprovarInput struct {
	OrcamentoID     string
	ClienteID       string
	OrdemServicoID  string
	FornecedorID    string
	IdempotencySeed string
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
	input.OrdemServicoID = strings.TrimSpace(input.OrdemServicoID)
	input.FornecedorID = strings.TrimSpace(input.FornecedorID)

	if !validation.IsUUID(input.OrcamentoID) {
		return domain.Aprovacao{}, ErrIdentificadorInvalido
	}
	if !validation.IsUUID(input.ClienteID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	if input.OrdemServicoID != "" && !validation.IsUUID(input.OrdemServicoID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	if !validation.IsUUID(input.FornecedorID) {
		return domain.Aprovacao{}, ErrFornecedorIDInvalido
	}
	input.IdempotencySeed = input.OrcamentoID + ":" + input.ClienteID
	return useCase.repository.Aprovar(ctx, input)
}
