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

type entregarRepositoryStub struct {
	resultado domain.ResultadoEntrega
	err       error
}

func (stub entregarRepositoryStub) Entregar(context.Context, application.EntregarInput) (domain.ResultadoEntrega, error) {
	return stub.resultado, stub.err
}

func TestEntregarHandler(t *testing.T) {
	const osID = "70000000-0000-0000-0000-000000000001"
	sucesso := domain.ResultadoEntrega{
		OrdemServicoID: osID, Status: domain.StatusEntregue, ValorFinal: 270,
		ResponsavelEntregaID: "90000000-0000-0000-0000-000000000001",
		ClienteID:            "20000000-0000-0000-0000-000000000001", DataEntrega: time.Now(),
	}
	cases := []struct {
		nome, id, corpo string
		err             error
		status          int
	}{
		{"osId invalido", "abc", `{}`, nil, http.StatusBadRequest},
		{"json invalido", osID, `{`, nil, http.StatusBadRequest},
		{"campo desconhecido", osID, `{"extra":true}`, nil, http.StatusBadRequest},
		{"cliente invalido", osID, `{"clienteId":"abc"}`, nil, http.StatusBadRequest},
		{"os inexistente", osID, `{}`, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"cliente inexistente", osID, `{}`, application.ErrClienteNaoEncontrado, http.StatusNotFound},
		{"os nao finalizada", osID, `{}`, domain.ErrOSNaoFinalizada, http.StatusConflict},
		{"os ja entregue", osID, `{}`, domain.ErrOSJaEntregue, http.StatusConflict},
		{"valor indisponivel", osID, `{}`, domain.ErrValorFinalIndisponivel, http.StatusConflict},
		{"sucesso", osID, `{"observacoes":"sem ressalvas"}`, nil, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.nome, func(t *testing.T) {
			useCase := application.NewEntregar(entregarRepositoryStub{resultado: sucesso, err: tc.err}, nil, nil)
			mux := http.NewServeMux()
			mux.Handle("POST /ordens-servico/{osId}/entrega", NewEntregarHandler(useCase))
			request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+tc.id+"/entrega", strings.NewReader(tc.corpo))
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != tc.status {
				t.Fatalf("status=%d, esperado=%d, body=%s", writer.Code, tc.status, writer.Body.String())
			}
			if tc.status == http.StatusOK {
				var response entregarResponse
				if err := json.Unmarshal(writer.Body.Bytes(), &response); err != nil || response.Status != domain.StatusEntregue || response.ValorFinal != 270 {
					t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
				}
			}
		})
	}
}
