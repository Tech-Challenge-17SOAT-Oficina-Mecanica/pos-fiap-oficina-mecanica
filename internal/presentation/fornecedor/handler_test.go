package fornecedor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type repositoryStub struct {
	fornecedor domain.Fornecedor
	filtros    application.FiltrosConsulta
	version    int
	err        error
}

func (stub repositoryStub) Cadastrar(_ context.Context, _ domain.Cadastro) (domain.Fornecedor, error) {
	return stub.fornecedor, stub.err
}

func (stub *repositoryStub) Listar(_ context.Context, filtros application.FiltrosConsulta) ([]domain.Fornecedor, int, error) {
	stub.filtros = filtros
	return []domain.Fornecedor{}, 0, stub.err
}

func (stub repositoryStub) BuscarPorID(_ context.Context, _ string) (domain.Fornecedor, error) {
	return stub.fornecedor, stub.err
}

func (stub *repositoryStub) Atualizar(_ context.Context, _ string, _ domain.Atualizacao, version int, _ string) (domain.Fornecedor, error) {
	stub.version = version
	return stub.fornecedor, stub.err
}

func TestCadastrarHandlerCriaFornecedor(t *testing.T) {
	useCase := application.NewCadastrar(repositoryStub{fornecedor: domain.Fornecedor{
		ID: "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa",
		Cadastro: domain.Cadastro{
			RazaoSocial:      "Auto Pecas Brasil Ltda",
			Documento:        "04252011000110",
			TipoDocumento:    "CNPJ",
			Telefone:         "11999990000",
			PrazoEntregaDias: 7,
		},
		Ativo:    true,
		Version:  1,
		CriadoEm: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC),
	}})

	request := httptest.NewRequest(http.MethodPost, "/fornecedores", bytes.NewBufferString(`{
		"razaoSocial":"Auto Pecas Brasil Ltda",
		"documento":"04.252.011/0001-10",
		"tipoDocumento":"CNPJ",
		"telefone":"11999990000"
	}`))
	response := httptest.NewRecorder()

	NewCadastrarHandler(useCase).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content-type=%q", response.Header().Get("Content-Type"))
	}
}

func TestCadastrarHandlerRejeitaDocumentoDuplicado(t *testing.T) {
	useCase := application.NewCadastrar(repositoryStub{err: application.ErrDocumentoDuplicado})
	request := httptest.NewRequest(http.MethodPost, "/fornecedores", bytes.NewBufferString(`{
		"razaoSocial":"Auto Pecas Brasil Ltda",
		"documento":"04.252.011/0001-10",
		"tipoDocumento":"CNPJ",
		"email":"vendas@example.com"
	}`))
	response := httptest.NewRecorder()

	NewCadastrarHandler(useCase).ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestListarHandlerAplicaTamanhoPadrao(t *testing.T) {
	repository := &repositoryStub{}
	useCase := application.NewConsultarFornecedores(repository)
	request := httptest.NewRequest(http.MethodGet, "/fornecedores", nil)
	response := httptest.NewRecorder()

	NewListarHandler(useCase).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if repository.filtros.Tamanho != application.TamanhoPaginaPadrao {
		t.Fatalf("tamanho=%d", repository.filtros.Tamanho)
	}
}

func TestListarHandlerRejeitaTamanhoAcimaDoMaximo(t *testing.T) {
	useCase := application.NewConsultarFornecedores(&repositoryStub{})
	request := httptest.NewRequest(http.MethodGet, "/fornecedores?tamanho=51", nil)
	response := httptest.NewRecorder()

	NewListarHandler(useCase).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAtualizarHandlerAtualizaFornecedor(t *testing.T) {
	repository := &repositoryStub{fornecedor: domain.Fornecedor{
		ID: "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa",
		Cadastro: domain.Cadastro{
			RazaoSocial:      "Auto Pecas Brasil Ltda",
			Documento:        "04252011000110",
			TipoDocumento:    "CNPJ",
			Telefone:         "11999990000",
			PrazoEntregaDias: 10,
		},
		Ativo:        true,
		Version:      2,
		AtualizadoEm: time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC),
	}}
	request := httptest.NewRequest(http.MethodPut, "/fornecedores/e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa", bytes.NewBufferString(`{
		"razaoSocial":"Auto Pecas Brasil Ltda",
		"telefone":"11999990000",
		"prazoEntregaDias":10
	}`))
	request.SetPathValue("fornecedorId", "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa")
	request.Header.Set("If-Match", "1")
	response := httptest.NewRecorder()

	NewAtualizarHandler(application.NewAtualizarFornecedor(repository)).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if repository.version != 1 {
		t.Fatalf("version=%d", repository.version)
	}
	var body fornecedorAtualizadoResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Version != 2 || body.DataAtualizacao == "" {
		t.Fatalf("body=%+v", body)
	}
}

func TestAtualizarHandlerExigeIfMatch(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/fornecedores/e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa", bytes.NewBufferString(`{}`))
	request.SetPathValue("fornecedorId", "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa")
	response := httptest.NewRecorder()

	NewAtualizarHandler(application.NewAtualizarFornecedor(&repositoryStub{})).ServeHTTP(response, request)

	if response.Code != http.StatusPreconditionRequired {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAtualizarHandlerRejeitaCamposImutaveis(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/fornecedores/e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa", bytes.NewBufferString(`{
		"razaoSocial":"Auto Pecas Brasil Ltda",
		"documento":"04252011000110",
		"telefone":"11999990000"
	}`))
	request.SetPathValue("fornecedorId", "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa")
	request.Header.Set("If-Match", "1")
	response := httptest.NewRecorder()

	NewAtualizarHandler(application.NewAtualizarFornecedor(&repositoryStub{})).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAtualizarHandlerMapeiaVersaoDivergente(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/fornecedores/e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa", bytes.NewBufferString(`{
		"razaoSocial":"Auto Pecas Brasil Ltda",
		"telefone":"11999990000"
	}`))
	request.SetPathValue("fornecedorId", "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa")
	request.Header.Set("If-Match", "1")
	response := httptest.NewRecorder()

	NewAtualizarHandler(application.NewAtualizarFornecedor(&repositoryStub{err: application.ErrVersaoDivergente})).ServeHTTP(response, request)

	if response.Code != http.StatusPreconditionFailed {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAtualizarHandlerMapeiaFornecedorInativo(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/fornecedores/e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa", bytes.NewBufferString(`{
		"razaoSocial":"Auto Pecas Brasil Ltda",
		"telefone":"11999990000"
	}`))
	request.SetPathValue("fornecedorId", "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa")
	request.Header.Set("If-Match", "1")
	response := httptest.NewRecorder()

	NewAtualizarHandler(application.NewAtualizarFornecedor(&repositoryStub{err: application.ErrFornecedorInativo})).ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAtualizarHandlerRejeitaAtualizacaoSemContato(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/fornecedores/e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa", bytes.NewBufferString(`{"razaoSocial":"Auto Pecas Brasil Ltda"}`))
	request.SetPathValue("fornecedorId", "e5d3e1bc-74f2-49ad-b3d8-95eed4ef0cfa")
	request.Header.Set("If-Match", "1")
	response := httptest.NewRecorder()

	NewAtualizarHandler(application.NewAtualizarFornecedor(&repositoryStub{err: errors.New("nao deveria chamar repositorio")})).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
