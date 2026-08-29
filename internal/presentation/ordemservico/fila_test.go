package ordemservico

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	security "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type filaRepositoryStub struct {
	input application.ConsultarFilaInput
	itens []domain.ItemFila
	total int
}

func (stub *filaRepositoryStub) ConsultarFila(_ context.Context, limite, deslocamento int) ([]domain.ItemFila, int, error) {
	stub.input = application.ConsultarFilaInput{Limite: limite, Deslocamento: deslocamento}
	return stub.itens, stub.total, nil
}

type filaTokenStub struct{ scopes []string }

func (stub filaTokenStub) Validar(_ string) (security.Claims, error) {
	return security.Claims{Escopos: stub.scopes}, nil
}

func TestConsultarFilaRetornaListaPaginada(t *testing.T) {
	mecanicoID := "11111111-1111-1111-1111-111111111111"
	entrada := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)
	repository := &filaRepositoryStub{total: 3, itens: []domain.ItemFila{{
		OrdemServicoID: "os-1", Placa: "ABC1D23", Marca: "Marca", Modelo: "Modelo",
		Status: "AGUARDANDO_EXECUCAO", MecanicoResponsavelID: &mecanicoID, DataEntradaFila: entrada,
	}}}
	handler := NewConsultarFilaHandler(application.NewConsultarFila(repository))
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodGet, "/fila-atendimento?pagina=1&tamanho=2", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; esperado 200; corpo=%s", response.Code, response.Body.String())
	}
	if repository.input.Limite != 2 || repository.input.Deslocamento != 2 {
		t.Fatalf("paginacao repassada = %+v", repository.input)
	}
	var body sharedhttp.Lista[filaItemResponse]
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TotalElementos != 3 || body.TotalPaginas != 2 || len(body.Data) != 1 || body.Data[0].MecanicoResponsavelID == nil {
		t.Fatalf("resposta inesperada: %+v", body)
	}
}

func TestConsultarFilaValidaPaginacao(t *testing.T) {
	repository := &filaRepositoryStub{}
	response := httptest.NewRecorder()
	NewConsultarFilaHandler(application.NewConsultarFila(repository))(
		response, httptest.NewRequest(http.MethodGet, "/fila-atendimento?tamanho=51", nil),
	)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; esperado 400", response.Code)
	}
}

func TestConsultarFilaExigeEscopoOSLer(t *testing.T) {
	repository := &filaRepositoryStub{}
	handler := seguranca.RequireScope(filaTokenStub{scopes: []string{"os:escrever"}}, "os:ler",
		NewConsultarFilaHandler(application.NewConsultarFila(repository)))
	request := httptest.NewRequest(http.MethodGet, "/fila-atendimento", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d; esperado 403", response.Code)
	}
}
