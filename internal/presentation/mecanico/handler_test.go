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
)

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
	cases := []struct {
		name   string
		body   string
		use    *cadastrarUseCase
		status int
	}{
		{"json invalido", `{`, &cadastrarUseCase{}, 400},
		{"dados invalidos", `{}`, &cadastrarUseCase{err: domain.ErrSenhaCurta}, 400},
		{"email duplicado", `{}`, &cadastrarUseCase{err: application.ErrEmailDuplicado}, 409},
		{"erro interno", `{}`, &cadastrarUseCase{err: errors.New("db")}, 500},
		{"sucesso", `{"nome":"Maria","email":"maria@oficina.local","senha":"senha-com-15-xxx","escopos":["clientes:ler"]}`, &cadastrarUseCase{mecanico: domain.Mecanico{ID: "m1", Nome: "Maria", Email: "maria@oficina.local", Ativo: true, Escopos: []string{"clientes:ler"}}}, 201},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			NewCadastrarHandler(test.use)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
		})
	}
}
