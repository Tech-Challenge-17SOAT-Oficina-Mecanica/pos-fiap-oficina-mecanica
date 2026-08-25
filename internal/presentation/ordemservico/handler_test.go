package ordemservico

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

type criarFake struct {
	ordem domain.OrdemDeServico
	err   error
}

func (fake criarFake) Execute(context.Context, application.CriarInput) (domain.OrdemDeServico, error) {
	return fake.ordem, fake.err
}

type tokenFake struct {
	claims seguranca.Claims
	err    error
}

func (fake tokenFake) Validar(string) (seguranca.Claims, error) { return fake.claims, fake.err }

func TestCriarHandler(t *testing.T) {
	validToken := tokenFake{claims: seguranca.Claims{Escopos: []string{escopoCriarOrdemServico}}}
	validUseCase := criarFake{ordem: domain.OrdemDeServico{
		ID: "os", ClienteID: "cliente", VeiculoID: "veiculo", Status: domain.StatusRecebida,
		CriadaEm: time.Date(2026, 8, 22, 10, 30, 0, 0, time.FixedZone("BRT", -3*60*60)),
	}}
	body := `{"clienteId":"00000000-0000-0000-0000-000000000001","veiculoId":"00000000-0000-0000-0000-000000000002"}`
	cases := []struct {
		name    string
		body    string
		auth    string
		useCase criarFake
		token   tokenFake
		status  int
		out     string
	}{
		{"sem token", body, "", validUseCase, validToken, http.StatusUnauthorized, ""},
		{"token inválido", body, "Bearer token", validUseCase, tokenFake{err: errors.New("jwt")}, http.StatusUnauthorized, ""},
		{"sem escopo", body, "Bearer token", validUseCase, tokenFake{claims: seguranca.Claims{Escopos: []string{"os:ler"}}}, http.StatusForbidden, ""},
		{"json inválido", `{`, "Bearer token", validUseCase, validToken, http.StatusBadRequest, ""},
		{"campo desconhecido", `{"clienteId":"id","veiculoId":"id","problema":"x"}`, "Bearer token", validUseCase, validToken, http.StatusBadRequest, ""},
		{"id inválido", body, "Bearer token", criarFake{err: application.ErrClienteIDInvalido}, validToken, http.StatusBadRequest, ""},
		{"cliente inexistente", body, "Bearer token", criarFake{err: application.ErrClienteNaoEncontrado}, validToken, http.StatusNotFound, ""},
		{"veículo inexistente", body, "Bearer token", criarFake{err: application.ErrVeiculoNaoEncontrado}, validToken, http.StatusNotFound, ""},
		{"vínculo inválido", body, "Bearer token", criarFake{err: application.ErrVeiculoNaoVinculadoCliente}, validToken, http.StatusConflict, ""},
		{"erro interno", body, "Bearer token", criarFake{err: errors.New("db")}, validToken, http.StatusInternalServerError, ""},
		{"sucesso", body, "Bearer token", validUseCase, validToken, http.StatusCreated, `"status":"RECEBIDA"`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(test.body))
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			NewCriarHandler(test.useCase, test.token)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.out != "" && !strings.Contains(response.Body.String(), test.out) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}
