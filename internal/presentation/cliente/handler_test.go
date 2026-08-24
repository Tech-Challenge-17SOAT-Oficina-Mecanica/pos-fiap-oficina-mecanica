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

type consultarFake struct {
	cliente domain.Cliente
	err     error
}

func (fake consultarFake) Execute(context.Context, string) (domain.Cliente, error) {
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

func TestConsultarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "usuario", Escopos: []string{escopoConsultarCliente}}}
	validUseCase := consultarFake{cliente: domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Version: 4, Veiculos: []domain.Veiculo{{ID: "v1", Placa: "ABC1D23", Marca: "Toyota", Modelo: "Corolla", Ano: 2020}}}}
	cases := []struct {
		name    string
		auth    string
		useCase consultarFake
		token   tokenFake
		status  int
		body    string
	}{
		{"sem token", "", validUseCase, validToken, http.StatusUnauthorized, ""},
		{"token invalido", "Bearer invalido", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized, ""},
		{"sem escopo", "Bearer jwt", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"clientes:escrever"}}}, http.StatusForbidden, ""},
		{"dominio invalido", "Bearer jwt", consultarFake{err: domain.ErrDocumentoInvalido}, validToken, http.StatusBadRequest, ""},
		{"nao encontrado", "Bearer jwt", consultarFake{err: application.ErrClienteNaoEncontrado}, validToken, http.StatusNotFound, ""},
		{"erro interno", "Bearer jwt", consultarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError, ""},
		{"sucesso", "Bearer jwt", validUseCase, validToken, http.StatusOK, `"veiculos":[{`},
		{"sem veiculo", "Bearer jwt", consultarFake{cliente: domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Version: 1}}, validToken, http.StatusOK, `"veiculos":[]`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/clientes?documento=39053344705", nil)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewConsultarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}
