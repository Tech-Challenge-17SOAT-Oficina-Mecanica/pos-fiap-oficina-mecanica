package ordemservico

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type listarRepositoryStub struct {
	itens []domain.ItemListagem
	total int
	err   error
}

func (stub listarRepositoryStub) Listar(context.Context, domain.FiltrosListagem, int, int) ([]domain.ItemListagem, int, error) {
	return stub.itens, stub.total, stub.err
}

func TestListarHandler(t *testing.T) {
	itemExemplo := domain.ItemListagem{
		OrdemServicoID: "70000000-0000-0000-0000-000000000001", Status: "RECEBIDA",
		ClienteID: "cliente-1", ClienteNome: "Ana", ClienteDocumento: "39053344705",
		VeiculoID: "veiculo-1", Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno",
	}

	t.Run("sem filtros retorna a lista", func(t *testing.T) {
		handler := NewListarHandler(application.NewListar(listarRepositoryStub{itens: []domain.ItemListagem{itemExemplo}, total: 1}))
		writer := httptest.NewRecorder()
		handler(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico", nil))
		if writer.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
		var resposta struct {
			Data           []listagemItemResponse `json:"data"`
			TotalElementos int                    `json:"totalElementos"`
		}
		if err := json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || len(resposta.Data) != 1 || resposta.TotalElementos != 1 {
			t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
		}
		if resposta.Data[0].Cliente.Documento != "39053344705" {
			t.Fatalf("cliente=%+v", resposta.Data[0].Cliente)
		}
	})

	t.Run("sem resultado retorna lista vazia, nao 404", func(t *testing.T) {
		handler := NewListarHandler(application.NewListar(listarRepositoryStub{itens: []domain.ItemListagem{}, total: 0}))
		writer := httptest.NewRecorder()
		handler(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico?status=CANCELADA", nil))
		if writer.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("status invalido", func(t *testing.T) {
		handler := NewListarHandler(application.NewListar(listarRepositoryStub{}))
		writer := httptest.NewRecorder()
		handler(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico?status=INEXISTENTE", nil))
		if writer.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("paginacao invalida", func(t *testing.T) {
		handler := NewListarHandler(application.NewListar(listarRepositoryStub{}))
		writer := httptest.NewRecorder()
		handler(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico?tamanho=999", nil))
		if writer.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("documento invalido", func(t *testing.T) {
		handler := NewListarHandler(application.NewListar(listarRepositoryStub{}))
		writer := httptest.NewRecorder()
		handler(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico?documento=123", nil))
		if writer.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("placa invalida", func(t *testing.T) {
		handler := NewListarHandler(application.NewListar(listarRepositoryStub{}))
		writer := httptest.NewRecorder()
		handler(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico?placa=###", nil))
		if writer.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})
}
