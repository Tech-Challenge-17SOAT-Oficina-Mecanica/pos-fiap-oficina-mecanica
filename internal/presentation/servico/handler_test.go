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
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type cadastrarFake struct {
	servico domain.Servico
	err     error
}

func (fake cadastrarFake) Execute(context.Context, domain.NovoServicoInput, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

type tokenFake struct {
	claims seguranca.Claims
	err    error
}

func (fake tokenFake) Validar(string) (seguranca.Claims, error) { return fake.claims, fake.err }

func TestCadastrarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "usuario", Escopos: []string{escopoCadastrarServico}}}
	validUseCase := cadastrarFake{servico: domain.Servico{
		ID: "id", Codigo: "SER-000004", Nome: "Revisão", Valor: "100.00", TempoEstimadoMinutos: 30,
		Ativo: true, Version: 1, DataCriacao: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
	}}
	validBody := `{"nome":"Revisão","valor":100,"tempoEstimadoMinutos":30}`
	cases := []struct {
		name, body, auth string
		useCase          cadastrarFake
		token            tokenFake
		status           int
	}{
		{"sem token", validBody, "", validUseCase, validToken, http.StatusUnauthorized},
		{"token inválido", validBody, "Bearer inválido", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized},
		{"sem escopo", validBody, "Bearer jwt", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"servicos:ler"}}}, http.StatusForbidden},
		{"json inválido", `{`, "Bearer jwt", validUseCase, validToken, http.StatusBadRequest},
		{"campo desconhecido", `{"nome":"Teste","valor":1,"tempoEstimadoMinutos":1,"extra":true}`, "Bearer jwt", validUseCase, validToken, http.StatusBadRequest},
		{"valor ausente", `{"nome":"Teste","tempoEstimadoMinutos":1}`, "Bearer jwt", validUseCase, validToken, http.StatusBadRequest},
		{"nome vazio", validBody, "Bearer jwt", cadastrarFake{err: domain.ErrNomeObrigatorio}, validToken, http.StatusBadRequest},
		{"duplicado", validBody, "Bearer jwt", cadastrarFake{err: application.ErrServicoDuplicado}, validToken, http.StatusConflict},
		{"erro interno", validBody, "Bearer jwt", cadastrarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError},
		{"sucesso", validBody, "Bearer jwt", validUseCase, validToken, http.StatusCreated},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/servicos", strings.NewReader(test.body))
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewCadastrarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.status == http.StatusCreated && !strings.Contains(response.Body.String(), `"codigo":"SER-000004"`) {
				t.Fatalf("resposta: %s", response.Body.String())
			}
		})
	}
}
