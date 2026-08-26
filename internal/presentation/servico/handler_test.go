package servico

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type cadastrarFake struct {
	servico domain.Servico
	err     error
}

func (fake cadastrarFake) Execute(context.Context, domain.NovoServicoInput, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

type consultarFake struct {
	pagina  application.Pagina
	servico domain.Servico
	err     error
}

type atualizarFake struct {
	servico domain.Servico
	err     error
}

type desativarFake struct {
	servico domain.Servico
	err     error
}

func (fake desativarFake) Execute(context.Context, string, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

type reativarFake struct {
	servico domain.Servico
	err     error
}

func (fake reativarFake) Execute(context.Context, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

func (fake atualizarFake) Execute(context.Context, string, int, domain.Atualizacao, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

func (fake consultarFake) Listar(context.Context, application.Filtros) (application.Pagina, error) {
	return fake.pagina, fake.err
}
func (fake consultarFake) PorID(context.Context, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

func TestCadastrarHandler(t *testing.T) {
	valid := cadastrarFake{servico: domain.Servico{ID: "id", Codigo: "SER-000004", Nome: "Revisão", Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true, Version: 1, DataCriacao: time.Now()}}
	cases := []struct {
		nome, body, campo string
		useCase           cadastrarFake
		status            int
	}{
		{"body inválido", `{`, "", valid, 400},
		{"valor ausente", `{"nome":"Teste","tempoEstimadoMinutos":1}`, "valor", valid, 400},
		{"nome inválido", `{"nome":"Teste","valor":1,"tempoEstimadoMinutos":1}`, "nome", cadastrarFake{err: domain.ErrNomeObrigatorio}, 400},
		{"conflito", `{"nome":"Teste","valor":1,"tempoEstimadoMinutos":1}`, "nome", cadastrarFake{err: application.ErrServicoDuplicado}, 409},
		{"sucesso", `{"nome":"Teste","valor":1,"tempoEstimadoMinutos":1}`, "", valid, 201},
	}
	for _, test := range cases {
		t.Run(test.nome, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/servicos", strings.NewReader(test.body))
			NewCadastrarHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.campo != "" && !strings.Contains(response.Body.String(), `"campo":"`+test.campo+`"`) {
				t.Fatalf("campo ausente: %s", response.Body.String())
			}
		})
	}
}

func TestListarHandler(t *testing.T) {
	valid := consultarFake{pagina: application.Pagina{Servicos: []domain.Servico{{ID: "id", Codigo: "SER-000001", Nome: "Revisão", Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true}}, Tamanho: 20, TotalElementos: 1, TotalPaginas: 1}}
	cases := []struct {
		nome, url, campo string
		useCase          consultarFake
		status           int
	}{
		{"página inválida", "/servicos?pagina=x", "pagina", valid, 400},
		{"booleano inválido", "/servicos?incluirInativos=x", "incluirInativos", valid, 400},
		{"tamanho fora do limite", "/servicos?tamanho=51", "tamanho", consultarFake{err: application.ErrTamanhoInvalido}, 400},
		{"sucesso", "/servicos", "", valid, 200},
	}
	for _, test := range cases {
		t.Run(test.nome, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.url, nil)
			NewListarHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.campo != "" && !strings.Contains(response.Body.String(), `"campo":"`+test.campo+`"`) {
				t.Fatalf("campo ausente: %s", response.Body.String())
			}
		})
	}
}

func TestConsultarHandler(t *testing.T) {
	const id = "40000000-0000-0000-0000-000000000001"
	valid := consultarFake{servico: domain.Servico{ID: id, Codigo: "SER-000001", Nome: "Revisão", Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true, Version: 3}}
	cases := []struct {
		nome, id, campo string
		useCase         consultarFake
		status          int
	}{
		{"uuid inválido", "abc", "servicoId", valid, 400},
		{"não encontrado", id, "", consultarFake{err: application.ErrServicoNaoEncontrado}, 404},
		{"erro interno", id, "", consultarFake{err: errors.New("db")}, 500},
		{"sucesso", id, "", valid, 200},
	}
	for _, test := range cases {
		t.Run(test.nome, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/servicos/"+test.id, nil)
			request.SetPathValue("servicoId", test.id)
			NewConsultarHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.campo != "" && !strings.Contains(response.Body.String(), `"campo":"`+test.campo+`"`) {
				t.Fatalf("campo ausente: %s", response.Body.String())
			}
		})
	}
}

func TestAtualizarHandler(t *testing.T) {
	const id = "40000000-0000-0000-0000-000000000001"
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	validUseCase := atualizarFake{servico: domain.Servico{
		ID: id, Codigo: "SER-000001", Nome: "Revisão completa", Valor: "180.00",
		TempoEstimadoMinutos: 40, Ativo: true, Version: 2, DataAtualizacao: &now,
	}}
	cases := []struct {
		name, id, body, ifMatch string
		useCase                 atualizarFake
		status                  int
		responseBody            string
	}{
		{"uuid inválido", "abc", `{"nome":"Revisão"}`, "1", validUseCase, http.StatusBadRequest, ""},
		{"sem If-Match", id, `{"nome":"Revisão"}`, "", validUseCase, http.StatusPreconditionRequired, ""},
		{"If-Match inválido", id, `{"nome":"Revisão"}`, "x", validUseCase, http.StatusBadRequest, ""},
		{"json inválido", id, `{`, "1", validUseCase, http.StatusBadRequest, ""},
		{"campo imutável", id, `{"codigo":"SER-999999"}`, "1", validUseCase, http.StatusBadRequest, ""},
		{"campo desconhecido", id, `{"extra":true}`, "1", validUseCase, http.StatusBadRequest, ""},
		{"dados inválidos", id, `{}`, "1", atualizarFake{err: domain.ErrAtualizacaoVazia}, http.StatusBadRequest, ""},
		{"duplicado", id, `{"nome":"Alinhamento"}`, "1", atualizarFake{err: application.ErrServicoDuplicado}, http.StatusConflict, ""},
		{"não encontrado", id, `{"nome":"Revisão"}`, "1", atualizarFake{err: application.ErrServicoNaoEncontrado}, http.StatusNotFound, ""},
		{"versão divergente", id, `{"nome":"Revisão"}`, "1", atualizarFake{err: application.ErrVersaoDivergente}, http.StatusPreconditionFailed, ""},
		{"erro interno", id, `{"nome":"Revisão"}`, "1", atualizarFake{err: errors.New("db")}, http.StatusInternalServerError, ""},
		{"sucesso", id, `{"nome":"Revisão completa","valor":180.00}`, "1", validUseCase, http.StatusOK, `"version":2`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPatch, "/servicos/"+test.id, strings.NewReader(test.body))
			request.SetPathValue("servicoId", test.id)
			if test.ifMatch != "" {
				request.Header.Set("If-Match", test.ifMatch)
			}
			response := httptest.NewRecorder()
			NewAtualizarHandler(test.useCase)(response, request)
			if response.Code != test.status || test.responseBody != "" && !strings.Contains(response.Body.String(), test.responseBody) {
				t.Fatalf("status %d, body: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestDesativarHandler(t *testing.T) {
	const id = "40000000-0000-0000-0000-000000000001"
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	valid := desativarFake{servico: domain.Servico{ID: id, Codigo: "SER-000001", Nome: "Revisão",
		Ativo: false, DataDesativacao: &now, UsuarioDesativacao: "usuario"}}
	cases := []struct {
		name, id string
		useCase  desativarFake
		status   int
		body     string
	}{
		{"uuid inválido", "abc", valid, http.StatusBadRequest, ""},
		{"não encontrado", id, desativarFake{err: application.ErrServicoNaoEncontrado}, http.StatusNotFound, ""},
		{"já inativo", id, desativarFake{err: domain.ErrServicoJaInativo}, http.StatusConflict, ""},
		{"sucesso", id, valid, http.StatusOK, `"ativo":false`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/servicos/"+test.id, nil)
			request.SetPathValue("servicoId", test.id)
			response := httptest.NewRecorder()
			NewDesativarHandler(test.useCase)(response, request)
			if response.Code != test.status || test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("status %d, body: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestReativarHandler(t *testing.T) {
	const id = "40000000-0000-0000-0000-000000000001"
	valid := reativarFake{servico: domain.Servico{ID: id, Codigo: "SER-000001", Nome: "Revisão", Ativo: true}}
	cases := []struct {
		name, id string
		useCase  reativarFake
		status   int
	}{
		{"uuid inválido", "abc", valid, http.StatusBadRequest},
		{"já ativo", id, reativarFake{err: domain.ErrServicoJaAtivo}, http.StatusConflict},
		{"nome duplicado", id, reativarFake{err: application.ErrServicoDuplicado}, http.StatusConflict},
		{"sucesso", id, valid, http.StatusOK},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/servicos/"+test.id+"/reativacao", nil)
			request.SetPathValue("servicoId", test.id)
			response := httptest.NewRecorder()
			NewReativarHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d, body: %s", response.Code, response.Body.String())
			}
		})
	}
}
