package seguranca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	"golang.org/x/crypto/bcrypt"
)

type repositoryFake struct {
	usuario domain.Usuario
	err     error
}

func (fake repositoryFake) BuscarPorEmail(context.Context, string) (domain.Usuario, error) {
	return fake.usuario, fake.err
}

type tokenFake struct {
	token string
	err   error
}

func (fake tokenFake) Gerar(string, []string) (string, error) { return fake.token, fake.err }

func TestAutenticar(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha"), bcrypt.MinCost)
	valid := domain.Usuario{ID: "id", SenhaHash: string(hash), Ativo: true, Escopos: []string{"veiculos:ler"}}
	cases := []struct {
		name, email, senha string
		repository         repositoryFake
		token              tokenFake
		want               error
		wantToken          string
	}{
		{"dados obrigatorios", "", "", repositoryFake{}, tokenFake{}, ErrDadosInvalidos, ""},
		{"usuario ausente", "a@b.com", "senha", repositoryFake{err: errors.New("ausente")}, tokenFake{}, ErrCredenciaisInvalidas, ""},
		{"usuario inativo", "a@b.com", "senha", repositoryFake{usuario: domain.Usuario{Ativo: false}}, tokenFake{}, ErrCredenciaisInvalidas, ""},
		{"senha invalida", "a@b.com", "outra", repositoryFake{usuario: valid}, tokenFake{}, ErrCredenciaisInvalidas, ""},
		{"falha no token", "a@b.com", "senha", repositoryFake{usuario: valid}, tokenFake{err: errors.New("token")}, errors.New("token"), ""},
		{"sucesso", " a@b.com ", " senha ", repositoryFake{usuario: valid}, tokenFake{token: "jwt"}, nil, "jwt"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewAutenticar(test.repository, test.token).Execute(context.Background(), test.email, test.senha)
			if test.want != nil && (err == nil || err.Error() != test.want.Error()) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && err != nil {
				t.Fatal(err)
			}
			if got != test.wantToken {
				t.Fatalf("token: %q", got)
			}
		})
	}
}
