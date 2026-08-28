package ordemservico

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type finalizarRepositoryStub struct {
	resultado domain.ResultadoFinalizacao
	err       error
}

func (stub finalizarRepositoryStub) Finalizar(context.Context, application.FinalizarInput) (domain.ResultadoFinalizacao, error) {
	return stub.resultado, stub.err
}

func TestFinalizarHandler(t *testing.T) {
	const osID = "70000000-0000-0000-0000-000000000001"
	sucesso := domain.ResultadoFinalizacao{OrdemServicoID: osID, Status: domain.StatusFinalizada, DataFinalizacao: time.Now(), Observacoes: "tudo ok"}

	casos := []struct {
		nome, id, corpo string
		err             error
		status          int
	}{
		{"osId invalido", "abc", `{}`, nil, http.StatusBadRequest},
		{"json invalido", osID, `{`, nil, http.StatusBadRequest},
		{"campo desconhecido", osID, `{"observacoes":"ok","extra":true}`, nil, http.StatusBadRequest},
		{"os inexistente", osID, `{}`, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"fora de execucao", osID, `{}`, domain.ErrOSNaoEmExecucao, http.StatusConflict},
		{"servicos pendentes", osID, `{}`, domain.ErrServicosPendentes, http.StatusConflict},
		{"complementar pendente", osID, `{}`, domain.ErrOrcamentoComplementarPendente, http.StatusConflict},
		{"reservas pendentes", osID, `{}`, domain.ErroReservasPendentes{Itens: []domain.ItemPendenteBaixa{{ItemID: "item-1", Codigo: "PEC-1", Quantidade: 2}}}, http.StatusConflict},
		{"sucesso", osID, `{"observacoes":"tudo ok"}`, nil, http.StatusOK},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			useCase := application.NewFinalizar(finalizarRepositoryStub{resultado: sucesso, err: caso.err})
			mux := http.NewServeMux()
			mux.Handle("POST /ordens-servico/{osId}/finalizar", NewFinalizarHandler(useCase))
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+caso.id+"/finalizar", strings.NewReader(caso.corpo))
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != caso.status {
				t.Fatalf("status=%d, esperado %d, body=%s", writer.Code, caso.status, writer.Body.String())
			}
			if caso.status == http.StatusOK {
				var resposta finalizarResponse
				if err := json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta.Status != domain.StatusFinalizada {
					t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
				}
			}
		})
	}
}
