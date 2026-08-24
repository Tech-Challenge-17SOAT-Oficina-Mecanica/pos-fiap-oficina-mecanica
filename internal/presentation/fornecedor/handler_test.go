package fornecedor

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type repositoryStub struct {
	fornecedor domain.Fornecedor
	err        error
}

func (stub repositoryStub) Cadastrar(_ context.Context, _ domain.Cadastro) (domain.Fornecedor, error) {
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
