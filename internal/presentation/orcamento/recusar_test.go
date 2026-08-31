package orcamento

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

type recusarRepositoryStub struct {
	result domain.Decisao
	err    error
}

func (stub recusarRepositoryStub) Recusar(_ context.Context, _ application.RecusarInput) (domain.Decisao, error) {
	return stub.result, stub.err
}

func TestRecusarHandler(t *testing.T) {
	const osID = "10000000-0000-0000-0000-000000000001"
	const orcamentoID = "20000000-0000-0000-0000-000000000001"
	jwt, err := infrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, _ := jwt.Gerar("mecanico", []string{"orcamentos:decidir"})
	cliente, _ := jwt.GerarCliente("cliente", osID)
	clienteOutraOS, _ := jwt.GerarCliente("cliente", "20000000-0000-0000-0000-000000000001")
	decisao := domain.Decisao{OrcamentoID: orcamentoID, OrdemServicoID: osID, TipoOrcamento: "PRINCIPAL", StatusOrcamento: "RECUSADO", StatusOrdemServico: "CANCELADA", ClienteID: "cliente", DecididoEm: time.Now(), Motivo: "valor alto"}

	tests := []struct {
		name, id, token, body string
		err                   error
		want                  int
	}{
		{"sem token", orcamentoID, "", "", nil, http.StatusUnauthorized},
		{"id invalido", "invalido", mecanico, "", nil, http.StatusBadRequest},
		{"cliente outra os", orcamentoID, clienteOutraOS, "", nil, http.StatusOK},
		{"corpo invalido", orcamentoID, mecanico, "{invalido", nil, http.StatusBadRequest},
		{"nao encontrado", orcamentoID, mecanico, "", application.ErrOrcamentoNaoEncontrado, http.StatusNotFound},
		{"acesso negado", orcamentoID, mecanico, "", application.ErrAcessoNegado, http.StatusForbidden},
		{"ja decidido", orcamentoID, mecanico, "", application.ErrOrcamentoJaDecidido, http.StatusConflict},
		{"complementar sem principal", orcamentoID, mecanico, "", application.ErrOrcamentoComplementarSemPrincipal, http.StatusConflict},
		{"os fora de aguardando aprovacao", orcamentoID, mecanico, "", application.ErrOrdemServicoNaoAguardandoAprovacao, http.StatusConflict},
		{"motivo invalido", orcamentoID, mecanico, "", application.ErrMotivoInvalido, http.StatusBadRequest},
		{"erro interno", orcamentoID, mecanico, "", errors.New("falhou"), http.StatusInternalServerError},
		{"mecanico", orcamentoID, mecanico, `{"motivo":"valor alto"}`, nil, http.StatusOK},
		{"cliente", orcamentoID, cliente, "", nil, http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			useCase := application.NewRecusar(recusarRepositoryStub{result: decisao, err: test.err}, nil, nil)
			mux.Handle("POST /orcamentos/{orcamentoId}/recusar", seguranca.RequireScope(jwt, "orcamentos:decidir", NewRecusarHandler(useCase)))
			var body *strings.Reader
			if test.body != "" {
				body = strings.NewReader(test.body)
			} else {
				body = strings.NewReader("")
			}
			request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+test.id+"/recusar", body)
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
		})
	}
}
