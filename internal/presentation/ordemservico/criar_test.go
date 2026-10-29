package ordemservico

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type criarUseCaseStub struct {
	input application.CriarInput
	ordem domain.OrdemDeServico
	err   error
}

func (stub *criarUseCaseStub) Execute(_ context.Context, input application.CriarInput) (domain.OrdemDeServico, error) {
	stub.input = input
	return stub.ordem, stub.err
}

func TestCriarHandlerRepassaServicosEPecas(t *testing.T) {
	criadaEm := time.Date(2026, 10, 28, 10, 30, 0, 0, time.UTC)
	stub := &criarUseCaseStub{ordem: domain.OrdemDeServico{ID: "os-1", ClienteID: "cliente", VeiculoID: "veiculo", Status: domain.StatusRecebida, CriadaEm: criadaEm}}
	body := `{"clienteId":"cliente","veiculoId":"veiculo",
		"servicos":[{"servicoId":"servico-1","observacao":"urgente"}],
		"pecas":[{"pecaId":"peca-1","quantidade":2}]}`
	response := httptest.NewRecorder()
	NewCriarHandler(stub).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(body)))

	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var created struct {
		OrdemServicoID string `json:"ordemServicoId"`
		Status         string `json:"status"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil || created.OrdemServicoID != "os-1" || created.Status != domain.StatusRecebida {
		t.Fatalf("resposta = %s, erro = %v", response.Body.String(), err)
	}
	servicos, pecas := stub.input.Servicos, stub.input.Pecas
	if len(servicos) != 1 || servicos[0].ServicoID != "servico-1" || servicos[0].Observacao != "urgente" {
		t.Fatalf("servicos = %+v", servicos)
	}
	if len(pecas) != 1 || pecas[0].ItemID != "peca-1" || pecas[0].Quantidade != 2 {
		t.Fatalf("pecas = %+v", pecas)
	}
}

func TestCriarHandlerRecusaCorpoInvalido(t *testing.T) {
	for _, body := range []string{`{"clienteId":"c","veiculoId":"v","itens":[]}`, `{"pecas":[{"itemId":"x"}]}`, `{`} {
		response := httptest.NewRecorder()
		NewCriarHandler(&criarUseCaseStub{}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(body)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("corpo %s: status %d", body, response.Code)
		}
	}
}

func TestCriarHandlerMapeiaErros(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		campo  string
	}{
		{"cliente invalido", application.ErrClienteIDInvalido, http.StatusBadRequest, "clienteId"},
		{"veiculo ausente", application.ErrVeiculoIDObrigatorio, http.StatusBadRequest, "veiculoId"},
		{"servico invalido", application.ErrServicoIDInvalido, http.StatusBadRequest, "servicos"},
		{"servico duplicado", domain.ErrServicoDuplicado, http.StatusBadRequest, "servicos"},
		{"peca invalida", application.ErrItemIDInvalido, http.StatusBadRequest, "pecas"},
		{"peca repetida", application.ErrItemRepetido, http.StatusBadRequest, "pecas"},
		{"insumo no lugar de peca", domainEstoque.ErrTipoItemInvalido, http.StatusBadRequest, "pecas"},
		{"quantidade fracionada", errors.New("a quantidade de peca deve ser inteira"), http.StatusBadRequest, "pecas"},
		{"cliente ausente", application.ErrClienteNaoEncontrado, http.StatusNotFound, ""},
		{"servico ausente", domain.ErrServicoNaoEncontrado, http.StatusNotFound, ""},
		{"peca ausente", application.ErrItemNaoEncontrado, http.StatusNotFound, ""},
		{"veiculo de outro cliente", application.ErrVeiculoNaoVinculadoCliente, http.StatusConflict, "veiculoId"},
		{"servico inativo", domain.ErrServicoInativo, http.StatusConflict, ""},
		{"peca inativa", application.ErrItemInativo, http.StatusConflict, ""},
		{"peca sem valor", application.ErrItemSemValor, http.StatusConflict, ""},
		{"inesperado", errors.New("db"), http.StatusInternalServerError, ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			body := `{"clienteId":"c","veiculoId":"v"}`
			NewCriarHandler(&criarUseCaseStub{err: test.err}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(body)))
			if response.Code != test.status {
				t.Fatalf("status %d, esperado %d", response.Code, test.status)
			}
			if test.campo != "" && !strings.Contains(response.Body.String(), `"campo":"`+test.campo+`"`) {
				t.Fatalf("campo %q ausente em %s", test.campo, response.Body.String())
			}
		})
	}
}
