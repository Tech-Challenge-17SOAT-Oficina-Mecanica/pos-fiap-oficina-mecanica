package orcamento

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

var ErrOrdemServicoNaoEncontrada = errors.New("ordem de servico nao encontrada")
var ErrOrcamentoNaoEncontrado = errors.New("orcamento nao encontrado para a ordem de servico")
var ErrAcessoNegado = errors.New("cliente sem acesso ao orcamento desta ordem de servico")

type Repository interface {
	Consultar(context.Context, string, string) (domain.Consulta, error)
}

type Consultar struct{ repository Repository }

func NewConsultar(repository Repository) Consultar { return Consultar{repository: repository} }

func (useCase Consultar) Execute(ctx context.Context, ordemServicoID, clienteID string) (domain.Consulta, error) {
	return useCase.repository.Consultar(ctx, ordemServicoID, clienteID)
}
