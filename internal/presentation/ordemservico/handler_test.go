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

type servicosHandlerRepositoryStub struct {
	resultado application.ResultadoRegistroServicos
	err       error
}

func (stub servicosHandlerRepositoryStub) RegistrarServicos(context.Context, string, []domain.ServicoCadastro) (application.ResultadoRegistroServicos, error) {
	return stub.resultado, stub.err
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

func TestRegistrarServicosHandler(t *testing.T) {
	const id = "10000000-0000-0000-0000-000000000001"
	ok := application.ResultadoRegistroServicos{
		Orcamento: domain.Orcamento{ID: "20000000-0000-0000-0000-000000000001", Tipo: domain.OrcamentoPrincipal, Status: domain.OrcamentoCriado, ValorTotal: 150},
		Servicos:  []domain.ServicoRegistrado{{ServicoID: id, Descricao: "Troca de oleo", ValorUnitario: 150, Observacao: "urgente"}},
	}
	tests := []struct {
		name, osID, body string
		err              error
		want             int
	}{
		{"os invalida", "invalido", `{}`, nil, http.StatusBadRequest},
		{"json invalido", id, `{`, nil, http.StatusBadRequest},
		{"campo desconhecido", id, `{"servicos":[],"x":true}`, nil, http.StatusBadRequest},
		{"lista vazia", id, `{"servicos":[]}`, nil, http.StatusBadRequest},
		{"servico invalido", id, `{"servicos":[{"servicoId":"invalido"}]}`, nil, http.StatusBadRequest},
		{"duplicado no corpo", id, `{"servicos":[{"servicoId":"` + id + `"},{"servicoId":"` + id + `"}]}`, nil, http.StatusBadRequest},
		{"os ausente", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"servico ausente", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, domain.ErrServicoNaoEncontrado, http.StatusNotFound},
		{"status invalido", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, domain.ErrStatusNaoPermiteServico, http.StatusConflict},
		{"orcamento ausente", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, domain.ErrOrcamentoAplicavelNaoEncontrado, http.StatusConflict},
		{"servico inativo", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, domain.ErrServicoInativo, http.StatusConflict},
		{"servico duplicado", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, domain.ErrServicoDuplicado, http.StatusConflict},
		{"erro interno", id, `{"servicos":[{"servicoId":"` + id + `"}]}`, errors.New("falhou"), http.StatusInternalServerError},
		{"sucesso", id, `{"servicos":[{"servicoId":"` + id + `","observacao":"urgente"}]}`, nil, http.StatusCreated},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewRegistrarServicosHandler(application.NewRegistrarServicos(servicosHandlerRepositoryStub{resultado: ok, err: test.err}))
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+test.osID+"/servicos", bytes.NewBufferString(test.body))
			request.SetPathValue("osId", test.osID)
			writer := httptest.NewRecorder()
			handler(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
			if test.want == http.StatusCreated && !bytes.Contains(writer.Body.Bytes(), []byte(`"servicoId"`)) {
				t.Fatalf("resposta=%s", writer.Body.String())
			}
		})
	}
}
