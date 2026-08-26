package servico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

var ErrServicoDuplicado = errors.New("já existe serviço ativo com o mesmo nome")

type Repository interface {
	ExisteAtivoPorNomeNormalizado(context.Context, string) (bool, error)
	Salvar(context.Context, domain.Servico) (domain.Servico, error)
}

type Cadastrar struct{ repository Repository }

func NewCadastrar(repository Repository) Cadastrar { return Cadastrar{repository: repository} }

func (useCase Cadastrar) Execute(ctx context.Context, input domain.NovoServicoInput, usuarioID string) (domain.Servico, error) {
	novo, err := domain.Novo(input)
	if err != nil {
		return domain.Servico{}, err
	}
	existe, err := useCase.repository.ExisteAtivoPorNomeNormalizado(ctx, novo.NomeNormalizado)
	if err != nil {
		return domain.Servico{}, err
	}
	if existe {
		return domain.Servico{}, ErrServicoDuplicado
	}
	novo.UsuarioCriacao = usuarioID
	return useCase.repository.Salvar(ctx, novo)
}
