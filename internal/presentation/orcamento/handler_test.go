package orcamento

import (
	"context"
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
	result application.Resultado
	err    error
}

func (fake useCaseFake) Execute(context.Context, application.ConsultarInput) (application.Resultado, error) {
	return fake.result, fake.err
}

type tokenFake struct {
	claims seguranca.Claims
	err    error
}

func (fake tokenFake) Validar(string) (seguranca.Claims, error) { return fake.claims, fake.err }

func TestConsultarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{Escopos: []string{escopoConsultarOrcamento}}}
	consulta := domain.Consulta{OrdemServicoID: "os", Orcamentos: []domain.Orcamento{}}
	cases := []struct {
		name, url, auth string
		useCase         useCaseFake
		token           tokenFake
		status          int
		contains        string
	}{
		{"sem token", "/orcamentos?documento=39053344705", "", useCaseFake{}, validToken, 401, "token"},
		{"token invalido", "/orcamentos?documento=39053344705", "Bearer jwt", useCaseFake{}, tokenFake{err: errors.New("jwt")}, 401, "token"},
		{"sem escopo", "/orcamentos?documento=39053344705", "Bearer jwt", useCaseFake{}, tokenFake{}, 403, "orcamentos:ler"},
		{"uuid invalido", "/orcamentos?orcamentoId=invalido", "Bearer jwt", useCaseFake{}, validToken, 400, "orcamentoId"},
		{"pagina invalida", "/orcamentos?documento=39053344705&pagina=x", "Bearer jwt", useCaseFake{}, validToken, 400, "pagina"},
		{"criterio ausente", "/orcamentos", "Bearer jwt", useCaseFake{err: application.ErrCriterioObrigatorio}, validToken, 400, "informe"},
		{"nao encontrado", "/orcamentos?orcamentoId=74000000-0000-0000-0000-000000000099", "Bearer jwt", useCaseFake{err: application.ErrOrcamentoNaoEncontrado}, validToken, 404, "encontrado"},
		{"erro interno", "/orcamentos?documento=39053344705", "Bearer jwt", useCaseFake{err: errors.New("db")}, validToken, 500, "falha"},
		{"por id", "/orcamentos?orcamentoId=74000000-0000-0000-0000-000000000001", "Bearer jwt", useCaseFake{result: application.Resultado{Consulta: &consulta}}, validToken, 200, `"ordemServicoId":"os"`},
		{"por documento", "/orcamentos?documento=39053344705", "Bearer jwt", useCaseFake{result: application.Resultado{Data: []domain.Consulta{consulta}, Pagina: 0, Tamanho: 20, TotalElementos: 1, TotalPaginas: 1}}, validToken, 200, `"totalElementos":1`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.url, nil)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewConsultarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.contains) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
