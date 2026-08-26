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

type atualizarUseCase struct {
	mecanico domain.Mecanico
	err      error
	input    application.AtualizarInput
}

func (fake *atualizarUseCase) Execute(_ context.Context, input application.AtualizarInput) (domain.Mecanico, error) {
	fake.input = input
	return fake.mecanico, fake.err
}

func (fake *cadastrarUseCase) Execute(_ context.Context, input domain.NovoMecanicoInput) (domain.Mecanico, error) {
	fake.input = input
	return fake.mecanico, fake.err
}

func TestAtualizarHandler(t *testing.T) {
	const id = "11111111-1111-1111-1111-111111111111"
	cases := []struct {
		name    string
		body    string
		use     *atualizarUseCase
		ifMatch string
		id      string
		status  int
	}{
		{"id invalido", `{}`, &atualizarUseCase{}, "1", "x", 400},
		{"sem if-match", `{}`, &atualizarUseCase{}, "", id, 428},
		{"if-match invalido", `{}`, &atualizarUseCase{}, "x", id, 400},
		{"json invalido", `{`, &atualizarUseCase{}, "1", id, 400},
		{"dados invalidos", `{}`, &atualizarUseCase{err: domain.ErrNomeObrigatorio}, "1", id, 400},
		{"nao encontrado", `{}`, &atualizarUseCase{err: application.ErrMecanicoNaoEncontrado}, "1", id, 404},
		{"email duplicado", `{}`, &atualizarUseCase{err: application.ErrEmailDuplicado}, "1", id, 409},
		{"versao", `{}`, &atualizarUseCase{err: application.ErrVersaoDivergente}, "1", id, 412},
		{"erro interno", `{}`, &atualizarUseCase{err: errors.New("db")}, "1", id, 500},
		{"sucesso", `{"nome":"Maria","email":"maria@oficina.local","escopos":["os:ler"]}`, &atualizarUseCase{mecanico: domain.Mecanico{ID: id, Nome: "Maria", Email: "maria@oficina.local", Ativo: true, Escopos: []string{"os:ler"}, Version: 2}}, "1", id, 200},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/mecanicos/"+test.id, strings.NewReader(test.body))
			request.SetPathValue("mecanicoId", test.id)
			if test.ifMatch != "" {
				request.Header.Set("If-Match", test.ifMatch)
			}
			response := httptest.NewRecorder()
			NewAtualizarHandler(test.use)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
		})
	}
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
