package ordemservico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var (
	ErrOrdemServicoForaDeRecebida   = errors.New("ordem de serviço não está com status RECEBIDA")
	ErrProblemaRelatadoJaRegistrado = errors.New("problema relatado já registrado")
)

type RegistrarProblemaRelatadoInput struct {
	OrdemServicoID string
	Descricao      string
	Observacoes    string
}

type ProblemaRelatadoRepository interface {
	RegistrarProblemaRelatado(context.Context, string, domain.ProblemaRelatado) (domain.OrdemDeServico, error)
}

type RegistrarProblemaRelatado struct{ repository ProblemaRelatadoRepository }

func NewRegistrarProblemaRelatado(repository ProblemaRelatadoRepository) RegistrarProblemaRelatado {
	return RegistrarProblemaRelatado{repository}
}

func (useCase RegistrarProblemaRelatado) Execute(ctx context.Context, input RegistrarProblemaRelatadoInput) (domain.OrdemDeServico, error) {
	problema, err := domain.NovoProblemaRelatado(input.Descricao, input.Observacoes)
	if err != nil {
		return domain.OrdemDeServico{}, err
	}
	return useCase.repository.RegistrarProblemaRelatado(ctx, input.OrdemServicoID, problema)
}
