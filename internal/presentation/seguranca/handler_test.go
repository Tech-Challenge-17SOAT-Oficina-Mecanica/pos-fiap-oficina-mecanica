package seguranca

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	"golang.org/x/crypto/bcrypt"
)

type repo struct {
	usuario domain.Usuario
	err     error
}

func (fake repo) BuscarPorEmail(context.Context, string) (domain.Usuario, error) {
	return fake.usuario, fake.err
}

type token struct {
	value string
	err   error
}

func (fake token) Gerar(string, []string) (string, error) { return fake.value, fake.err }

func TestLoginHandler(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("senha"), bcrypt.MinCost)
	useCase := application.NewAutenticar(repo{usuario: domain.Usuario{ID: "id", SenhaHash: string(hash), Ativo: true}}, token{value: "jwt"})
	handler := NewLoginHandler(useCase)
	for _, test := range []struct {
		body   string
		status int
	}{{"{", 400}, {`{"email":"","senha":""}`, 400}, {`{"email":"a@b.com","senha":"errada"}`, 401}, {`{"email":"a@b.com","senha":"senha"}`, 200}} {
		request := httptest.NewRequest(http.MethodPost, "/autenticacao/login", strings.NewReader(test.body))
		response := httptest.NewRecorder()
		handler(response, request)
		if response.Code != test.status {
			t.Fatalf("status %d", response.Code)
		}
	}
	failing := NewLoginHandler(application.NewAutenticar(repo{usuario: domain.Usuario{ID: "id", SenhaHash: string(hash), Ativo: true}}, token{err: errors.New("falha")}))
	response := httptest.NewRecorder()
	failing(response, httptest.NewRequest(http.MethodPost, "/autenticacao/login", strings.NewReader(`{"email":"a@b.com","senha":"senha"}`)))
	if response.Code != 500 {
		t.Fatalf("status %d", response.Code)
	}
}
