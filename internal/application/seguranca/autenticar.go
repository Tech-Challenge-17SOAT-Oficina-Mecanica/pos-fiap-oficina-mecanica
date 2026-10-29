package seguranca

import (
	"context"
	"errors"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
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

type VerificadorSenha interface {
	Verificar(hash, senha string) error
}

type Autenticar struct {
	senhas     VerificadorSenha
	repository UsuarioRepository
	token      TokenGenerator
}

func NewAutenticar(repository UsuarioRepository, token TokenGenerator, senhas VerificadorSenha) Autenticar {
	return Autenticar{repository: repository, token: token, senhas: senhas}
}

func (useCase Autenticar) Execute(ctx context.Context, email, senha string) (string, error) {
	email, senha = strings.TrimSpace(email), strings.TrimSpace(senha)
	if email == "" || senha == "" {
		return "", ErrDadosInvalidos
	}

	usuario, err := useCase.repository.BuscarPorEmail(ctx, email)
	if err != nil || !usuario.Ativo || useCase.senhas.Verificar(usuario.SenhaHash, senha) != nil {
		return "", ErrCredenciaisInvalidas
	}
	return useCase.token.Gerar(usuario.ID, usuario.Escopos)
}
