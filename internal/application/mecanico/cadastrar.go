package mecanico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailDuplicado = errors.New("e-mail já cadastrado")
var ErrMecanicoNaoEncontrado = errors.New("mecânico não encontrado")
var ErrVersaoDivergente = errors.New("If-Match divergente")

type MecanicoRepository interface {
	EmailExiste(context.Context, string) (bool, error)
	EmailExisteExcetoMecanico(context.Context, string, string) (bool, error)
	BuscarPorID(context.Context, string) (domain.Mecanico, error)
	SalvarMecanico(context.Context, domain.Mecanico, string) (domain.Mecanico, error)
	AtualizarMecanico(context.Context, domain.Mecanico, int) (domain.Mecanico, error)
}

type Cadastrar struct {
	repository MecanicoRepository
}

func NewCadastrar(repository MecanicoRepository) Cadastrar {
	return Cadastrar{repository: repository}
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
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return domain.Mecanico{}, err
	}
	return useCase.repository.SalvarMecanico(ctx, mecanico, string(hash))
}

type AtualizarInput struct {
	MecanicoID string
	Version    int
	Dados      domain.AtualizarMecanicoInput
}

type Atualizar struct {
	repository MecanicoRepository
}

func NewAtualizar(repository MecanicoRepository) Atualizar {
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
