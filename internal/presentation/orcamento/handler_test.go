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

type repositoryStub struct {
	result domain.Consulta
	err    error
}

func (stub repositoryStub) Consultar(context.Context, string, string) (domain.Consulta, error) {
	return stub.result, stub.err
}

func TestConsultarHandler(t *testing.T) {
	const osID = "10000000-0000-0000-0000-000000000001"
	jwt, err := infrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, _ := jwt.Gerar("mecanico", []string{"os:ler"})
	cliente, _ := jwt.GerarCliente("cliente", osID)
	clienteOutraOS, _ := jwt.GerarCliente("cliente", "20000000-0000-0000-0000-000000000001")
	estimativa := 3
	consulta := domain.Consulta{Cliente: domain.Cliente{ID: "cliente", Nome: "Ana", Documento: "123", TipoDocumento: "CPF"}, OrdemServicoID: osID, StatusOrdemServico: "EM_DIAGNOSTICO", Orcamentos: []domain.Orcamento{{ID: "orcamento", OriginalID: "original", Tipo: "PRINCIPAL", Status: "CRIADO", EstimativaDias: &estimativa, DataGeracao: time.Now(), Itens: []domain.Item{{Tipo: "SERVICO", Descricao: "Troca", Quantidade: 1, ValorUnitario: 100, ValorTotal: 100}}, Problemas: []domain.Problema{{ID: "problema", Descricao: "Ruido", RegistradoEm: time.Now()}}, ValorTotal: 100}}, ValorTotalGeral: 100}
	tests := []struct {
		name, id, token string
		err             error
		want            int
	}{
		{"sem token", osID, "", nil, http.StatusUnauthorized},
		{"id invalido", "invalido", mecanico, nil, http.StatusBadRequest},
		{"cliente outra os", osID, clienteOutraOS, nil, http.StatusForbidden},
		{"os ausente", osID, mecanico, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"orcamento ausente", osID, mecanico, application.ErrOrcamentoNaoEncontrado, http.StatusNotFound},
		{"acesso negado", osID, mecanico, application.ErrAcessoNegado, http.StatusForbidden},
		{"erro interno", osID, mecanico, errors.New("falhou"), http.StatusInternalServerError},
		{"mecanico", osID, mecanico, nil, http.StatusOK},
		{"cliente", osID, cliente, nil, http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			useCase := application.NewConsultar(repositoryStub{result: consulta, err: test.err})
			mux.Handle("GET /ordens-servico/{osId}/orcamento", seguranca.RequireAnyScope(jwt, []string{"os:ler", "orcamentos:ler"}, NewConsultarHandler(useCase)))
			request := httptest.NewRequest(http.MethodGet, "/ordens-servico/"+test.id+"/orcamento", nil)
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
			if test.want == http.StatusOK && (!strings.Contains(writer.Body.String(), `"documento":"***.***.***-23"`) || strings.Contains(writer.Body.String(), `"documento":"123"`)) {
				t.Fatalf("documento exposto: %s", writer.Body.String())
			}
		})
	}
}
