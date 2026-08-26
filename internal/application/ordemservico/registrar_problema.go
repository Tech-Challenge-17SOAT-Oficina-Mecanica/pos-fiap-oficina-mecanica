package ordemservico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var ErrOrdemServicoNaoEncontrada = errors.New("ordem de servico nao encontrada")

type ResultadoRegistroProblema struct {
	Problema  domain.Problema
	Orcamento domain.Orcamento
}

type ProblemaRepository interface {
	RegistrarProblema(context.Context, string, domain.ProblemaCadastro) (ResultadoRegistroProblema, error)
}

type RegistrarProblema struct{ repository ProblemaRepository }

func NewRegistrarProblema(repository ProblemaRepository) RegistrarProblema {
	return RegistrarProblema{repository: repository}
}

func (useCase RegistrarProblema) Execute(ctx context.Context, ordemServicoID string, cadastro domain.ProblemaCadastro) (ResultadoRegistroProblema, error) {
	return useCase.repository.RegistrarProblema(ctx, ordemServicoID, cadastro)
}
