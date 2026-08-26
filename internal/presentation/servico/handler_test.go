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

func TestCadastrarHandler(t *testing.T) {
	validUseCase := cadastrarFake{servico: domain.Servico{
		ID: "id", Codigo: "SER-000004", Nome: "Revisão", Valor: "100.00", TempoEstimadoMinutos: 30,
		Ativo: true, Version: 1, DataCriacao: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
	}}
	validBody := `{"nome":"Revisão","valor":100,"tempoEstimadoMinutos":30}`
	cases := []struct {
		name, body, campo string
		useCase           cadastrarFake
		status            int
	}{
		{"json inválido", `{`, "", validUseCase, http.StatusBadRequest},
		{"campo desconhecido", `{"nome":"Teste","valor":1,"tempoEstimadoMinutos":1,"extra":true}`, "", validUseCase, http.StatusBadRequest},
		{"valor ausente", `{"nome":"Teste","tempoEstimadoMinutos":1}`, "valor", validUseCase, http.StatusBadRequest},
		{"nome vazio", validBody, "nome", cadastrarFake{err: domain.ErrNomeObrigatorio}, http.StatusBadRequest},
		{"valor inválido", validBody, "valor", cadastrarFake{err: domain.ErrValorInvalido}, http.StatusBadRequest},
		{"tempo inválido", validBody, "tempoEstimadoMinutos", cadastrarFake{err: domain.ErrTempoEstimadoInvalido}, http.StatusBadRequest},
		{"duplicado", validBody, "nome", cadastrarFake{err: application.ErrServicoDuplicado}, http.StatusConflict},
		{"erro interno", validBody, "", cadastrarFake{err: errors.New("db")}, http.StatusInternalServerError},
		{"sucesso", validBody, "", validUseCase, http.StatusCreated},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/servicos", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			NewCadastrarHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.campo != "" && !strings.Contains(response.Body.String(), `"campo":"`+test.campo+`"`) {
				t.Fatalf("campo %q não informado: %s", test.campo, response.Body.String())
			}
			if test.status == http.StatusCreated && !strings.Contains(response.Body.String(), `"codigo":"SER-000004"`) {
				t.Fatalf("resposta: %s", response.Body.String())
			}
		})
	}
}
