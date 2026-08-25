package cliente

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

type atualizarFake struct {
	cliente domain.Cliente
	err     error
}

func (fake atualizarFake) Execute(context.Context, application.AtualizarInput) (domain.Cliente, error) {
	return fake.cliente, fake.err
}

type inativarFake struct {
	result application.Inativacao
	err    error
}

func (fake inativarFake) Execute(context.Context, application.InativarInput) (application.Inativacao, error) {
	return fake.result, fake.err
}

type reativarFake struct {
	result application.Reativacao
	err    error
}

func (fake reativarFake) Execute(context.Context, application.ReativarInput) (application.Reativacao, error) {
	return fake.result, fake.err
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
	validUseCase := consultarFake{cliente: domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 4, Veiculos: []domain.Veiculo{{ID: "v1", Placa: "ABC1D23", Marca: "Toyota", Modelo: "Corolla", Ano: 2020}}}}
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
		{"sucesso", "Bearer jwt", validUseCase, validToken, http.StatusOK, `"ativo":true`},
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

func TestAtualizarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "usuario", Escopos: []string{escopoCadastrarCliente}}}
	validUseCase := atualizarFake{cliente: domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 3}}
	validBody := `{"nome":"Ana","documento":"39053344705","tipoDocumento":"CPF","telefone":"11988887777"}`
	cases := []struct {
		name    string
		body    string
		auth    string
		ifMatch string
		useCase atualizarFake
		token   tokenFake
		status  int
		bodyOut string
	}{
		{"sem token", validBody, "", "2", validUseCase, validToken, http.StatusUnauthorized, ""},
		{"token invalido", validBody, "Bearer invalido", "2", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized, ""},
		{"sem escopo", validBody, "Bearer jwt", "2", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"clientes:ler"}}}, http.StatusForbidden, ""},
		{"sem if match", validBody, "Bearer jwt", "", validUseCase, validToken, http.StatusPreconditionRequired, ""},
		{"if match invalido", validBody, "Bearer jwt", "abc", validUseCase, validToken, http.StatusBadRequest, ""},
		{"json invalido", `{`, "Bearer jwt", "2", validUseCase, validToken, http.StatusBadRequest, ""},
		{"dominio invalido", validBody, "Bearer jwt", "2", atualizarFake{err: domain.ErrNomeObrigatorio}, validToken, http.StatusBadRequest, ""},
		{"nao encontrado", validBody, "Bearer jwt", "2", atualizarFake{err: application.ErrClienteNaoEncontrado}, validToken, http.StatusNotFound, ""},
		{"duplicado", validBody, "Bearer jwt", "2", atualizarFake{err: application.ErrClienteDuplicado}, validToken, http.StatusConflict, ""},
		{"versao divergente", validBody, "Bearer jwt", "2", atualizarFake{err: application.ErrVersaoDivergente}, validToken, http.StatusPreconditionFailed, ""},
		{"erro interno", validBody, "Bearer jwt", "2", atualizarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError, ""},
		{"sucesso", validBody, "Bearer jwt", "2", validUseCase, validToken, http.StatusOK, `"ativo":true`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/clientes/id", strings.NewReader(test.body))
			request.SetPathValue("clienteId", "id")
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			if test.ifMatch != "" {
				request.Header.Set("If-Match", test.ifMatch)
			}
			response := httptest.NewRecorder()
			NewAtualizarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.bodyOut != "" && !strings.Contains(response.Body.String(), test.bodyOut) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}

func TestInativarHandler(t *testing.T) {
	now := time.Now()
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "00000000-0000-0000-0000-000000000001", Escopos: []string{escopoCadastrarCliente}}}
	validUseCase := inativarFake{result: application.Inativacao{
		Cliente:            domain.Cliente{ID: "00000000-0000-0000-0000-000000000002", Nome: "Ana", Ativo: false, InativadoEm: &now, InativadoPor: "00000000-0000-0000-0000-000000000001", Motivo: "duplicado"},
		VeiculosInativados: []domain.VeiculoInativado{{ID: "v1", Placa: "ABC1D23"}},
		DocumentoLiberado:  true,
	}}
	cases := []struct {
		name    string
		auth    string
		id      string
		useCase inativarFake
		token   tokenFake
		status  int
		body    string
	}{
		{"sem token", "", "00000000-0000-0000-0000-000000000002", validUseCase, validToken, http.StatusUnauthorized, ""},
		{"token invalido", "Bearer invalido", "00000000-0000-0000-0000-000000000002", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized, ""},
		{"sem escopo", "Bearer jwt", "00000000-0000-0000-0000-000000000002", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"clientes:ler"}}}, http.StatusForbidden, ""},
		{"uuid invalido", "Bearer jwt", "id", validUseCase, validToken, http.StatusBadRequest, ""},
		{"motivo invalido", "Bearer jwt", "00000000-0000-0000-0000-000000000002", inativarFake{err: domain.ErrMotivoInvalido}, validToken, http.StatusBadRequest, ""},
		{"nao encontrado", "Bearer jwt", "00000000-0000-0000-0000-000000000002", inativarFake{err: application.ErrClienteNaoEncontrado}, validToken, http.StatusNotFound, ""},
		{"ja inativo", "Bearer jwt", "00000000-0000-0000-0000-000000000002", inativarFake{err: application.ErrClienteJaInativo}, validToken, http.StatusNoContent, ""},
		{"os aberta", "Bearer jwt", "00000000-0000-0000-0000-000000000002", inativarFake{err: application.OSAbertaError{Ordens: []application.OrdemServicoAberta{{ID: "os", Status: "EM_EXECUCAO"}}}}, validToken, http.StatusConflict, `"ordemServicoId":"os"`},
		{"erro interno", "Bearer jwt", "00000000-0000-0000-0000-000000000002", inativarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError, ""},
		{"sucesso", "Bearer jwt", "00000000-0000-0000-0000-000000000002", validUseCase, validToken, http.StatusOK, `"veiculosInativados":[{`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/clientes/"+test.id+"?motivo=duplicado", nil)
			request.SetPathValue("clienteId", test.id)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewInativarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}

func TestReativarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "00000000-0000-0000-0000-000000000001", Escopos: []string{escopoCadastrarCliente}}}
	validUseCase := reativarFake{result: application.Reativacao{Cliente: domain.Cliente{ID: "00000000-0000-0000-0000-000000000002", Nome: "Ana", Ativo: true}, ReativadoEm: time.Now()}}
	cases := []struct {
		name    string
		auth    string
		id      string
		useCase reativarFake
		token   tokenFake
		status  int
		body    string
	}{
		{"sem token", "", "00000000-0000-0000-0000-000000000002", validUseCase, validToken, http.StatusUnauthorized, ""},
		{"token invalido", "Bearer invalido", "00000000-0000-0000-0000-000000000002", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized, ""},
		{"sem escopo", "Bearer jwt", "00000000-0000-0000-0000-000000000002", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"clientes:ler"}}}, http.StatusForbidden, ""},
		{"uuid invalido", "Bearer jwt", "id", validUseCase, validToken, http.StatusBadRequest, ""},
		{"nao encontrado", "Bearer jwt", "00000000-0000-0000-0000-000000000002", reativarFake{err: application.ErrClienteNaoEncontrado}, validToken, http.StatusNotFound, ""},
		{"duplicado", "Bearer jwt", "00000000-0000-0000-0000-000000000002", reativarFake{err: application.ErrClienteDuplicado}, validToken, http.StatusConflict, ""},
		{"ja ativo", "Bearer jwt", "00000000-0000-0000-0000-000000000002", reativarFake{err: application.ErrClienteJaAtivo}, validToken, http.StatusConflict, ""},
		{"erro interno", "Bearer jwt", "00000000-0000-0000-0000-000000000002", reativarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError, ""},
		{"sucesso", "Bearer jwt", "00000000-0000-0000-0000-000000000002", validUseCase, validToken, http.StatusOK, `"veiculosReativados":0`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/clientes/"+test.id+"/reativacao", nil)
			request.SetPathValue("clienteId", test.id)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewReativarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}
