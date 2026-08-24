package cliente

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

type cadastrarFake struct {
	cliente domain.Cliente
	err     error
}

func (fake cadastrarFake) Execute(context.Context, domain.NovoClienteInput) (domain.Cliente, error) {
	return fake.cliente, fake.err
}

type tokenFake struct {
	claims seguranca.Claims
	err    error
}

func (fake tokenFake) Validar(string) (seguranca.Claims, error) {
	return fake.claims, fake.err
}

func TestCadastrarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "usuario", Escopos: []string{escopoCadastrarCliente}}}
	validUseCase := cadastrarFake{cliente: domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777"}}
	cases := []struct {
		name    string
		body    string
		auth    string
		useCase cadastrarFake
		token   tokenFake
		status  int
	}{
		{"sem token", `{}`, "", validUseCase, validToken, http.StatusUnauthorized},
		{"token invalido", `{}`, "Bearer invalido", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized},
		{"sem escopo", `{}`, "Bearer jwt", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"clientes:ler"}}}, http.StatusForbidden},
		{"json invalido", `{`, "Bearer jwt", validUseCase, validToken, http.StatusBadRequest},
		{"dominio invalido", `{}`, "Bearer jwt", cadastrarFake{err: domain.ErrNomeObrigatorio}, validToken, http.StatusBadRequest},
		{"duplicado", `{}`, "Bearer jwt", cadastrarFake{err: application.ErrClienteDuplicado}, validToken, http.StatusConflict},
		{"erro interno", `{}`, "Bearer jwt", cadastrarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError},
		{"sucesso", `{}`, "Bearer jwt", validUseCase, validToken, http.StatusCreated},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(test.body))
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewCadastrarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
		})
	}
}
