package seguranca

import (
	"context"
	"errors"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDadosInvalidos       = errors.New("email e senha são obrigatórios")
	ErrCredenciaisInvalidas = errors.New("credenciais inválidas")
)

type UsuarioRepository interface {
	BuscarPorEmail(context.Context, string) (seguranca.Usuario, error)
}

type TokenGenerator interface {
	Gerar(string, []string) (string, error)
}

type Autenticar struct {
	repository UsuarioRepository
	token      TokenGenerator
}

func NewAutenticar(repository UsuarioRepository, token TokenGenerator) Autenticar {
	return Autenticar{repository: repository, token: token}
}

func (useCase Autenticar) Execute(ctx context.Context, email, senha string) (string, error) {
	email, senha = strings.TrimSpace(email), strings.TrimSpace(senha)
	if email == "" || senha == "" {
		return "", ErrDadosInvalidos
	}

	usuario, err := useCase.repository.BuscarPorEmail(ctx, email)
	if err != nil || !usuario.Ativo || bcrypt.CompareHashAndPassword([]byte(usuario.SenhaHash), []byte(senha)) != nil {
		return "", ErrCredenciaisInvalidas
	}
	return useCase.token.Gerar(usuario.ID, usuario.Escopos)
}
