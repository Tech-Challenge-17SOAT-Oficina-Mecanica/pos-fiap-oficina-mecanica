package mecanico

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailDuplicado = errors.New("e-mail já cadastrado")

type MecanicoRepository interface {
	EmailExiste(context.Context, string) (bool, error)
	SalvarMecanico(context.Context, domain.Mecanico, string) (domain.Mecanico, error)
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
