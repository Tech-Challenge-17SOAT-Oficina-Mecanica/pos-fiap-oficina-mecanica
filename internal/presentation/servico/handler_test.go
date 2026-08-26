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
