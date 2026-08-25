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

type useCaseFake struct {
	result domain.OrdemDeServico
	err    error
	input  application.RegistrarProblemaRelatadoInput
}

func (fake *useCaseFake) Execute(_ context.Context, input application.RegistrarProblemaRelatadoInput) (domain.OrdemDeServico, error) {
	fake.input = input
	return fake.result, fake.err
}

func TestRegistrarProblemaRelatadoHandler(t *testing.T) {
	osID := "70000000-0000-0000-0000-000000000001"
	success := &useCaseFake{result: domain.OrdemDeServico{
		ID: osID, Status: domain.StatusEmDiagnostico,
		ProblemaRelatado:      domain.ProblemaRelatado{Descricao: "Ruído", Observacoes: "Há uma semana"},
		DataInicioDiagnostico: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC),
	}}
	cases := []struct {
		name    string
		id      string
		body    string
		useCase *useCaseFake
		status  int
		out     string
	}{
		{"osId inválido", "id", `{}`, success, http.StatusBadRequest, `"campo":"osId"`},
		{"json inválido", osID, `{`, success, http.StatusBadRequest, ""},
		{"campo desconhecido", osID, `{"descricao":"Ruído","extra":true}`, success, http.StatusBadRequest, ""},
		{"descrição vazia", osID, `{"descricao":""}`, &useCaseFake{err: domain.ErrDescricaoObrigatoria}, http.StatusBadRequest, `"campo":"descricao"`},
		{"OS inexistente", osID, `{"descricao":"Ruído"}`, &useCaseFake{err: application.ErrOrdemServicoNaoEncontrada}, http.StatusNotFound, ""},
		{"status inválido", osID, `{"descricao":"Ruído"}`, &useCaseFake{err: application.ErrOrdemServicoForaDeRecebida}, http.StatusConflict, ""},
		{"duplicado", osID, `{"descricao":"Ruído"}`, &useCaseFake{err: application.ErrProblemaRelatadoJaRegistrado}, http.StatusConflict, ""},
		{"erro interno", osID, `{"descricao":"Ruído"}`, &useCaseFake{err: errors.New("db")}, http.StatusInternalServerError, ""},
		{"sucesso", osID, `{"descricao":"Ruído","observacoes":"Há uma semana"}`, success, http.StatusCreated, `"status":"EM_DIAGNOSTICO"`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+test.id+"/problema-relatado", strings.NewReader(test.body))
			request.SetPathValue("osId", test.id)
			response := httptest.NewRecorder()
			NewRegistrarProblemaRelatadoHandler(test.useCase)(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.out != "" && !strings.Contains(response.Body.String(), test.out) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
	if success.input.OrdemServicoID != osID || success.input.Descricao != "Ruído" || success.input.Observacoes != "Há uma semana" {
		t.Fatalf("input: %+v", success.input)
	}
}
