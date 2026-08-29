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
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

type consultarRepositoryStub struct {
	resultado domain.ConsultaDetalhada
	err       error
	clienteID string
}

func (stub *consultarRepositoryStub) Consultar(_ context.Context, _ string, clienteID string) (domain.ConsultaDetalhada, error) {
	stub.clienteID = clienteID
	return stub.resultado, stub.err
}

func TestConsultarHandler(t *testing.T) {
	const osID = "70000000-0000-0000-0000-000000000001"
	jwt, err := seguranca.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, _ := jwt.Gerar("mecanico", []string{"os:ler"})
	cliente, _ := jwt.GerarCliente("cliente-1", osID)
	clienteOutraOS, _ := jwt.GerarCliente("cliente-1", "20000000-0000-0000-0000-000000000099")
	orcamentoReader, _ := jwt.Gerar("orcamento-reader", []string{"orcamentos:ler"})

	sucesso := domain.ConsultaDetalhada{
		OrdemServicoID: osID, StatusOrdemServico: "EM_EXECUCAO",
		Cliente: domain.ClienteResumo{ID: "cliente-1", Nome: "Ana", Documento: "12345678900"},
		Veiculo: domain.VeiculoResumo{ID: "veiculo-1", Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", Ano: 2020},
		Eventos: []domain.EventoConsulta{{
			ID:         "evt-1",
			Agregado:   "ORDEM_SERVICO",
			AgregadoID: osID,
			TipoEvento: "ORDEM_SERVICO_ATUALIZADA",
			Dados:      json.RawMessage(`{"statusAnterior":"AGUARDANDO_APROVACAO","statusNovo":"EM_EXECUCAO","clienteId":"cliente-1"}`),
			Metadados:  json.RawMessage(`{"usuarioId":"usuario-1"}`),
			OcorridoEm: time.Now(),
		}},
	}

	montar := func(stub *consultarRepositoryStub) http.Handler {
		handler := NewConsultarHandler(application.NewConsultar(stub))
		return segurancaPresentation.RequireScope(jwt, "os:ler", handler)
	}

	requisitar := func(handler http.Handler, id, token string) *httptest.ResponseRecorder {
		mux := http.NewServeMux()
		mux.Handle("GET /ordens-servico/{osId}", handler)
		request := httptest.NewRequest(http.MethodGet, "/ordens-servico/"+id, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		writer := httptest.NewRecorder()
		mux.ServeHTTP(writer, request)
		return writer
	}

	t.Run("osId invalido", func(t *testing.T) {
		writer := requisitar(montar(&consultarRepositoryStub{resultado: sucesso}), "invalido", mecanico)
		if writer.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("cliente de outra os", func(t *testing.T) {
		writer := requisitar(montar(&consultarRepositoryStub{resultado: sucesso}), osID, clienteOutraOS)
		if writer.Code != http.StatusForbidden {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("os inexistente", func(t *testing.T) {
		writer := requisitar(montar(&consultarRepositoryStub{err: application.ErrOrdemServicoNaoEncontrada}), osID, mecanico)
		if writer.Code != http.StatusNotFound {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("orcamentos:ler sem acesso", func(t *testing.T) {
		writer := requisitar(montar(&consultarRepositoryStub{resultado: sucesso}), osID, orcamentoReader)
		if writer.Code != http.StatusForbidden {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("sucesso via mecanico", func(t *testing.T) {
		stub := &consultarRepositoryStub{resultado: sucesso}
		writer := requisitar(montar(stub), osID, mecanico)
		if writer.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
		if stub.clienteID != "" {
			t.Fatalf("clienteID nao deveria ser repassado para token de mecanico: %q", stub.clienteID)
		}
		var resposta consultaResponse
		if err := json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta.Cliente.Nome != "Ana" || resposta.Veiculo.Placa != "ABC1D23" {
			t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
		}
		if resposta.Cliente.Documento != "***.***.***-00" {
			t.Fatalf("cliente.documento=%q, esperado mascarado", resposta.Cliente.Documento)
		}
		if len(resposta.Eventos) != 1 {
			t.Fatalf("eventos=%+v", resposta.Eventos)
		}
		if resposta.Eventos[0].StatusAnterior != "AGUARDANDO_APROVACAO" || resposta.Eventos[0].StatusNovo != "EM_EXECUCAO" {
			t.Fatalf("status transicao=%+v", resposta.Eventos[0])
		}
		if strings.Contains(writer.Body.String(), "\"dados\"") || strings.Contains(writer.Body.String(), "\"metadados\"") {
			t.Fatalf("resposta nao deve expor payload bruto da auditoria: %s", writer.Body.String())
		}
		if resposta.Problemas == nil || resposta.Orcamentos == nil || resposta.Eventos == nil {
			t.Fatalf("listas devem vir vazias, nao nulas: %+v", resposta)
		}
	})

	t.Run("sucesso via cliente", func(t *testing.T) {
		stub := &consultarRepositoryStub{resultado: sucesso}
		writer := requisitar(montar(stub), osID, cliente)
		if writer.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
		if stub.clienteID != "cliente-1" {
			t.Fatalf("clienteID=%q, esperado cliente-1", stub.clienteID)
		}
	})
}
