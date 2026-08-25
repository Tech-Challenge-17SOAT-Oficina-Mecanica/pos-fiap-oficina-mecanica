package orcamento

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

type useCaseFake struct {
	total json.Number
	err   error
}

func (fake useCaseFake) Execute(context.Context, string, string) (json.Number, error) {
	return fake.total, fake.err
}

type tokenFake struct {
	claims seguranca.Claims
	err    error
}

func (fake tokenFake) Validar(string) (seguranca.Claims, error) { return fake.claims, fake.err }

func TestCalcularHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{UsuarioID: "90000000-0000-0000-0000-000000000001", Escopos: []string{escopoCalcularOrcamento}}}
	validID := "74000000-0000-0000-0000-000000000002"
	cases := []struct {
		name, id, auth, contains string
		useCase                  useCaseFake
		token                    tokenFake
		status                   int
	}{
		{"sem token", validID, "", "token", useCaseFake{}, validToken, 401},
		{"token invalido", validID, "Bearer jwt", "token", useCaseFake{}, tokenFake{err: errors.New("jwt")}, 401},
		{"sem escopo", validID, "Bearer jwt", "orcamentos:escrever", useCaseFake{}, tokenFake{}, 403},
		{"id invalido", "invalido", "Bearer jwt", "orcamentoId", useCaseFake{}, validToken, 400},
		{"nao encontrado", validID, "Bearer jwt", "não encontrado", useCaseFake{err: application.ErrOrcamentoNaoEncontrado}, validToken, 404},
		{"item invalido", validID, "Bearer jwt", "item", useCaseFake{err: domain.ErrItemInvalido}, validToken, 400},
		{"conflito", validID, "Bearer jwt", "CRIADO", useCaseFake{err: domain.ErrStatusInvalido}, validToken, 409},
		{"interno", validID, "Bearer jwt", "falha", useCaseFake{err: errors.New("db")}, validToken, 500},
		{"sucesso", validID, "Bearer jwt", `"valorTotalGeral":754.00`, useCaseFake{total: "754.00"}, validToken, 200},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+test.id+"/calcular", nil)
			request.SetPathValue("orcamentoId", test.id)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewCalcularHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.contains) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
