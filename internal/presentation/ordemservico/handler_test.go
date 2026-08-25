package ordemservico

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type handlerRepositoryStub struct {
	resultado application.ResultadoRegistroProblema
	err       error
}

func (stub handlerRepositoryStub) RegistrarProblema(context.Context, string, domain.ProblemaCadastro) (application.ResultadoRegistroProblema, error) {
	return stub.resultado, stub.err
}

func TestRegistrarProblemaHandler(t *testing.T) {
	ok := application.ResultadoRegistroProblema{
		Problema:  domain.Problema{ID: "10000000-0000-0000-0000-000000000001", Descricao: "freio", Observacoes: "urgente"},
		Orcamento: domain.Orcamento{ID: "20000000-0000-0000-0000-000000000001", Tipo: domain.OrcamentoPrincipal, Status: domain.OrcamentoCriado},
	}
	tests := []struct {
		name, id, body string
		err            error
		want           int
	}{
		{"id invalido", "invalido", `{}`, nil, http.StatusBadRequest},
		{"json invalido", "10000000-0000-0000-0000-000000000001", `{`, nil, http.StatusBadRequest},
		{"campo desconhecido", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio","x":true}`, nil, http.StatusBadRequest},
		{"descricao vazia", "10000000-0000-0000-0000-000000000001", `{"descricao":" "}`, nil, http.StatusBadRequest},
		{"os ausente", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio"}`, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"status invalido", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio"}`, domain.ErrStatusNaoPermiteProblema, http.StatusConflict},
		{"orcamento fechado", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio"}`, domain.ErrOrcamentoFechado, http.StatusConflict},
		{"principal ausente", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio"}`, domain.ErrOrcamentoPrincipalNaoEncontrado, http.StatusConflict},
		{"erro interno", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio"}`, errors.New("falhou"), http.StatusInternalServerError},
		{"sucesso", "10000000-0000-0000-0000-000000000001", `{"descricao":"freio","observacoes":"urgente"}`, nil, http.StatusCreated},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewRegistrarProblemaHandler(application.NewRegistrarProblema(handlerRepositoryStub{resultado: ok, err: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+test.id+"/problemas", bytes.NewBufferString(test.body))
			request.SetPathValue("osId", test.id)
			writer := httptest.NewRecorder()
			handler(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
			if test.want == http.StatusCreated && !bytes.Contains(writer.Body.Bytes(), []byte(`"problemaId"`)) {
				t.Fatalf("resposta=%s", writer.Body.String())
			}
		})
	}
}
