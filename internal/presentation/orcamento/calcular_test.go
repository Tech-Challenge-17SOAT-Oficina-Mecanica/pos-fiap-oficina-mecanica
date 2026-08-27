package orcamento

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	orcamentoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

const orcamentoID = "74000000-0000-0000-0000-000000000001"

type calcularFake struct {
	alvo   orcamentoDomain.Orcamento
	erro   error
	irmaos []orcamento.OrcamentoDaOS
}

func (fake calcularFake) BuscarParaCalculo(context.Context, string) (orcamentoDomain.Orcamento, string, error) {
	return fake.alvo, "70000000-0000-0000-0000-000000000001", fake.erro
}

func (fake calcularFake) OrcamentosDaOrdem(context.Context, string) ([]orcamento.OrcamentoDaOS, error) {
	return fake.irmaos, nil
}

func (calcularFake) SalvarItens(context.Context, string, []orcamentoDomain.Item) error { return nil }

func calcular(t *testing.T, id string, fake calcularFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+id+"/calcular", nil)
	request.SetPathValue("orcamentoId", id)
	NewCalcularHandler(orcamento.NewCalcular(fake)).ServeHTTP(response, request)
	return response
}

func TestCalcularRetorna200(t *testing.T) {
	fake := calcularFake{alvo: orcamentoDomain.Orcamento{
		ID: orcamentoID, Tipo: orcamentoDomain.TipoPrincipal, Status: orcamentoDomain.StatusCriado,
		Itens: []orcamentoDomain.Item{{ID: "i1", Quantidade: 2, ValorUnitario: 45}},
	}}

	response := calcular(t, orcamentoID, fake)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d. corpo=%s", response.Code, response.Body.String())
	}

	var corpo map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}
	if corpo["valorTotalGeral"] != 90.00 {
		t.Fatalf("valorTotalGeral = %v, esperado 90", corpo["valorTotalGeral"])
	}
	// A estimativa fica fora desta entrega; não deve aparecer prometendo algo que não calcula.
	if _, presente := corpo["estimativaEntregaDias"]; presente {
		t.Fatal("estimativaEntregaDias não deveria estar na resposta ainda")
	}
}

func TestCalcularErros(t *testing.T) {
	casos := []struct {
		nome   string
		id     string
		fake   calcularFake
		status int
	}{
		{"id inválido", "nao-e-uuid", calcularFake{}, http.StatusBadRequest},
		{"não encontrado", orcamentoID, calcularFake{erro: orcamento.ErrOrcamentoNaoEncontrado}, http.StatusNotFound},
		{"já aprovado", orcamentoID, calcularFake{alvo: orcamentoDomain.Orcamento{Tipo: orcamentoDomain.TipoPrincipal, Status: orcamentoDomain.StatusAprovado}}, http.StatusConflict},
		{"complementar sem principal", orcamentoID, calcularFake{alvo: orcamentoDomain.Orcamento{Tipo: orcamentoDomain.TipoComplementar, Status: orcamentoDomain.StatusCriado}}, http.StatusConflict},
		{"complementar vinculado a outra OS", orcamentoID, calcularFake{alvo: orcamentoDomain.Orcamento{Tipo: orcamentoDomain.TipoComplementar, Status: orcamentoDomain.StatusCriado, OriginalID: "de-outra-os"}}, http.StatusConflict},
		{"sem itens", orcamentoID, calcularFake{alvo: orcamentoDomain.Orcamento{Tipo: orcamentoDomain.TipoPrincipal, Status: orcamentoDomain.StatusCriado}}, http.StatusConflict},
		{"falha inesperada", orcamentoID, calcularFake{erro: context.DeadlineExceeded}, http.StatusInternalServerError},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			response := calcular(t, caso.id, caso.fake)
			if response.Code != caso.status {
				t.Fatalf("status = %d, esperado %d. corpo=%s", response.Code, caso.status, response.Body.String())
			}
			if tipo := response.Header().Get("Content-Type"); tipo != "application/problem+json" {
				t.Fatalf("Content-Type = %q", tipo)
			}
		})
	}
}
