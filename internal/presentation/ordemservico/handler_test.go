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
)

type criarFake struct {
	ordem domain.OrdemDeServico
	err   error
}

func (fake criarFake) Execute(context.Context, application.CriarInput) (domain.OrdemDeServico, error) {
	return fake.ordem, fake.err
}

func TestCriarHandler(t *testing.T) {
	validUseCase := criarFake{ordem: domain.OrdemDeServico{
		ID: "os", ClienteID: "cliente", VeiculoID: "veiculo", Status: domain.StatusRecebida,
		CriadaEm: time.Date(2026, 8, 22, 10, 30, 0, 0, time.FixedZone("BRT", -3*60*60)),
	}}
	body := `{"clienteId":"00000000-0000-0000-0000-000000000001","veiculoId":"00000000-0000-0000-0000-000000000002"}`
	cases := []struct {
		name    string
		body    string
		useCase criarFake
		status  int
		out     string
	}{
		{"json inválido", `{`, validUseCase, http.StatusBadRequest, ""},
		{"campo desconhecido", `{"clienteId":"id","veiculoId":"id","problema":"x"}`, validUseCase, http.StatusBadRequest, ""},
		{"id inválido", body, criarFake{err: application.ErrClienteIDInvalido}, http.StatusBadRequest, ""},
		{"cliente inexistente", body, criarFake{err: application.ErrClienteNaoEncontrado}, http.StatusNotFound, ""},
		{"veículo inexistente", body, criarFake{err: application.ErrVeiculoNaoEncontrado}, http.StatusNotFound, ""},
		{"vínculo inválido", body, criarFake{err: application.ErrVeiculoNaoVinculadoCliente}, http.StatusConflict, ""},
		{"erro interno", body, criarFake{err: errors.New("db")}, http.StatusInternalServerError, ""},
		{"sucesso", body, validUseCase, http.StatusCreated, `"status":"RECEBIDA"`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			NewCriarHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.out != "" && !strings.Contains(response.Body.String(), test.out) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}
