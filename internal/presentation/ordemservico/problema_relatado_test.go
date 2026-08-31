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

type problemaRelatadoRepositoryStub struct {
	resultado domain.OrdemDeServico
	err       error
}

func (stub problemaRelatadoRepositoryStub) RegistrarProblemaRelatado(context.Context, string, domain.ProblemaRelatado) (domain.OrdemDeServico, error) {
	return stub.resultado, stub.err
}

func TestRegistrarProblemaRelatadoHandler(t *testing.T) {
	const osID = "70000000-0000-0000-0000-000000000001"
	inicio := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	sucesso := domain.OrdemDeServico{ID: osID, Status: domain.StatusEmDiagnostico,
		ProblemaRelatado: domain.ProblemaRelatado{Descricao: "Ruído", Observacoes: "Há uma semana"}, DataInicioDiagnostico: &inicio}
	casos := []struct {
		nome, id, corpo string
		err             error
		status          int
	}{
		{"osId inválido", "abc", `{}`, nil, http.StatusBadRequest},
		{"json inválido", osID, `{`, nil, http.StatusBadRequest},
		{"campo desconhecido", osID, `{"descricao":"Ruído","extra":true}`, nil, http.StatusBadRequest},
		{"descrição vazia", osID, `{"descricao":" "}`, nil, http.StatusBadRequest},
		{"OS inexistente", osID, `{"descricao":"Ruído"}`, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"status inválido", osID, `{"descricao":"Ruído"}`, application.ErrOrdemServicoForaDeRecebida, http.StatusConflict},
		{"duplicado", osID, `{"descricao":"Ruído"}`, application.ErrProblemaRelatadoJaRegistrado, http.StatusConflict},
		{"erro interno", osID, `{"descricao":"Ruído"}`, errors.New("db"), http.StatusInternalServerError},
		{"sucesso", osID, `{"descricao":"Ruído","observacoes":"Há uma semana"}`, nil, http.StatusCreated},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			handler := NewRegistrarProblemaRelatadoHandler(application.NewRegistrarProblemaRelatado(problemaRelatadoRepositoryStub{resultado: sucesso, err: caso.err}, nil, nil))
			requisicao := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+caso.id+"/problema-relatado", strings.NewReader(caso.corpo))
			requisicao.SetPathValue("osId", caso.id)
			resposta := httptest.NewRecorder()
			handler(resposta, requisicao)
			if resposta.Code != caso.status {
				t.Fatalf("status=%d body=%s", resposta.Code, resposta.Body.String())
			}
		})
	}
}
