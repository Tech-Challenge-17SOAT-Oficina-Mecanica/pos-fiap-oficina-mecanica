package mecanico

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

type tokenValidator struct {
	claims seguranca.Claims
	err    error
}

func (fake tokenValidator) Validar(string) (seguranca.Claims, error) { return fake.claims, fake.err }

type cadastrarUseCase struct {
	mecanico domain.Mecanico
	err      error
	input    domain.NovoMecanicoInput
}

func (fake *cadastrarUseCase) Execute(_ context.Context, input domain.NovoMecanicoInput) (domain.Mecanico, error) {
	fake.input = input
	return fake.mecanico, fake.err
}

func TestCadastrarHandler(t *testing.T) {
	autorizado := tokenValidator{claims: seguranca.Claims{UsuarioID: "u1", Escopos: []string{"mecanicos:escrever"}}}
	cases := []struct {
		name   string
		body   string
		token  tokenValidator
		use    *cadastrarUseCase
		header bool
		status int
	}{
		{"sem token", `{}`, autorizado, &cadastrarUseCase{}, false, 401},
		{"token invalido", `{}`, tokenValidator{err: errors.New("jwt")}, &cadastrarUseCase{}, true, 401},
		{"sem escopo", `{}`, tokenValidator{claims: seguranca.Claims{Escopos: []string{"clientes:ler"}}}, &cadastrarUseCase{}, true, 403},
		{"json invalido", `{`, autorizado, &cadastrarUseCase{}, true, 400},
		{"dados invalidos", `{}`, autorizado, &cadastrarUseCase{err: domain.ErrSenhaCurta}, true, 400},
		{"email duplicado", `{}`, autorizado, &cadastrarUseCase{err: application.ErrEmailDuplicado}, true, 409},
		{"erro interno", `{}`, autorizado, &cadastrarUseCase{err: errors.New("db")}, true, 500},
		{"sucesso", `{"nome":"Maria","email":"maria@oficina.local","senha":"senha-com-15-xxx","escopos":["clientes:ler"]}`, autorizado, &cadastrarUseCase{mecanico: domain.Mecanico{ID: "m1", Nome: "Maria", Email: "maria@oficina.local", Ativo: true, Escopos: []string{"clientes:ler"}}}, true, 201},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(test.body))
			if test.header {
				request.Header.Set("Authorization", "Bearer jwt")
			}
			response := httptest.NewRecorder()
			NewCadastrarHandler(test.use, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
		})
	}
}
