package mecanico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
)

var ErrEmailDuplicado = errors.New("e-mail já cadastrado")
var ErrMecanicoNaoEncontrado = errors.New("mecânico não encontrado")
var ErrVersaoDivergente = errors.New("If-Match divergente")

type GeradorHashSenha interface {
	Gerar(senha string) (string, error)
}

type Cadastrar struct {
	senhas     GeradorHashSenha
	repository CadastrarRepository
}

func NewCadastrar(repository CadastrarRepository, senhas GeradorHashSenha) Cadastrar {
	return Cadastrar{repository: repository, senhas: senhas}
}

func (useCase Cadastrar) Execute(ctx context.Context, input domain.NovoMecanicoInput) (domain.Mecanico, error) {
	mecanico, senha, err := domain.Novo(input)
	if err != nil {
		return domain.Mecanico{}, err
	}
	exists, err := useCase.repository.EmailExiste(ctx, mecanico.Email)
	if err != nil {
		return domain.Mecanico{}, err
	}
	if exists {
		return domain.Mecanico{}, ErrEmailDuplicado
	}
	hash, err := useCase.senhas.Gerar(senha)
	if err != nil {
		return domain.Mecanico{}, err
	}
	return useCase.repository.SalvarMecanico(ctx, mecanico, hash)
}

type AtualizarInput struct {
	MecanicoID string
	Version    int
	Dados      domain.AtualizarMecanicoInput
}

type Atualizar struct {
	repository AtualizarRepository
}

func NewAtualizar(repository AtualizarRepository) Atualizar {
	return Atualizar{repository: repository}
}

func (useCase Atualizar) Execute(ctx context.Context, input AtualizarInput) (domain.Mecanico, error) {
	atual, err := useCase.repository.BuscarPorID(ctx, input.MecanicoID)
	if err != nil {
		return domain.Mecanico{}, err
	}
	if atual.Version != input.Version {
		return domain.Mecanico{}, ErrVersaoDivergente
	}
	mecanico, err := atual.Atualizar(input.Dados)
	if err != nil {
		return domain.Mecanico{}, err
	}
	exists, err := useCase.repository.EmailExisteExcetoMecanico(ctx, mecanico.Email, mecanico.ID)
	if err != nil {
		return domain.Mecanico{}, err
	}
	if exists {
		return domain.Mecanico{}, ErrEmailDuplicado
	}
	return useCase.repository.AtualizarMecanico(ctx, mecanico, input.Version)
}
