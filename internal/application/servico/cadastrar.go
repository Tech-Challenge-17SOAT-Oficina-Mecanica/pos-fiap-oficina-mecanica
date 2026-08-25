package servico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

var ErrNomeDuplicado = errors.New("já existe serviço ativo com o mesmo nome normalizado")

type CadastrarRepository interface {
	Cadastrar(context.Context, domain.Servico) (domain.Servico, error)
}

type CadastrarServico struct {
	repository CadastrarRepository
}

func NewCadastrarServico(repository CadastrarRepository) CadastrarServico {
	return CadastrarServico{repository: repository}
}

func (uc CadastrarServico) Execute(ctx context.Context, input domain.NovoCadastroInput) (domain.Servico, error) {
	servico, err := domain.Novo(input)
	if err != nil {
		return domain.Servico{}, err
	}
	return uc.repository.Cadastrar(ctx, servico)
}
