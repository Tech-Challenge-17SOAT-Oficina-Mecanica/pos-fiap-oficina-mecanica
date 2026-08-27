package insumo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	insumoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

type consultarRepositorioFake struct {
	itens []insumoDomain.Insumo
	total int
	erro  error
}

func (fake consultarRepositorioFake) BuscarPorFiltro(context.Context, insumo.FiltrosConsulta, int, int) ([]insumoDomain.Insumo, int, error) {
	return fake.itens, fake.total, fake.erro
}

func (fake consultarRepositorioFake) BuscarPorID(context.Context, string) (insumoDomain.Insumo, error) {
	if len(fake.itens) == 0 {
		return insumoDomain.Insumo{}, insumo.ErrInsumoNaoEncontrado
	}
	return fake.itens[0], fake.erro
}

func insumoExemplo() insumoDomain.Insumo {
	custo := "32.00"
	return insumoDomain.Insumo{
		ID: "50000000-0000-0000-0000-000000000003", Codigo: "INS-000001",
		Nome: "Oleo 5W30", Descricao: "Oleo sintetico 5W30",
		CategoriaID: "10000000-0000-0000-0000-000000000002", Categoria: "Lubrificantes",
		UnidadeMedida: "L", CustoUnitario: &custo,
		SaldoFisico: "20.000", SaldoReservado: "2.000", EstoqueMinimo: "5.000",
		Ativo: true, Version: 1,
	}
}

func executarConsulta(t *testing.T, alvo string, fake consultarRepositorioFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler := NewConsultarInsumosHandler(insumo.NewConsultarInsumos(fake))
	handler(response, httptest.NewRequest(http.MethodGet, alvo, nil))
	return response
}

func TestConsultarInsumosRetornaEnvelopeCompleto(t *testing.T) {
	response := executarConsulta(t, "/estoque/insumos?codigo=INS-000001&quantidadeDesejada=18",
		consultarRepositorioFake{itens: []insumoDomain.Insumo{insumoExemplo()}, total: 1})
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo=%s", response.Code, response.Body.String())
	}

	var corpo struct {
		Data []struct {
			Codigo               string      `json:"codigo"`
			Tipo                 string      `json:"tipo"`
			CustoUnitario        json.Number `json:"custoUnitario"`
			SaldoDisponivel      json.Number `json:"saldoDisponivel"`
			QuantidadeDesejada   json.Number `json:"quantidadeDesejada"`
			QuantidadeDisponivel *bool       `json:"quantidadeDisponivel"`
			Disponivel           bool        `json:"disponivel"`
			AbaixoDoMinimo       bool        `json:"abaixoDoMinimo"`
			Version              int         `json:"version"`
		} `json:"data"`
		TotalElementos int `json:"totalElementos"`
		TotalPaginas   int `json:"totalPaginas"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}
	item := corpo.Data[0]
	if item.Tipo != "INSUMO" || item.Codigo != "INS-000001" || item.CustoUnitario.String() != "32.00" {
		t.Fatalf("identificação incorreta: %+v", item)
	}
	if item.SaldoDisponivel.String() != "18" || item.QuantidadeDesejada.String() != "18" {
		t.Fatalf("decimais incorretos: %+v", item)
	}
	if item.QuantidadeDisponivel == nil || !*item.QuantidadeDisponivel || !item.Disponivel || item.AbaixoDoMinimo {
		t.Fatalf("disponibilidade incorreta: %+v", item)
	}
	if corpo.TotalElementos != 1 || corpo.TotalPaginas != 1 {
		t.Fatalf("envelope incorreto: %+v", corpo)
	}
}

func TestConsultarInsumosListaVaziaRetorna200(t *testing.T) {
	response := executarConsulta(t, "/estoque/insumos?codigo=INS-999999", consultarRepositorioFake{})
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"data":[]`) {
		t.Fatalf("lista vazia deveria retornar 200: %d %s", response.Code, response.Body.String())
	}
}

func TestConsultarInsumosErrosDeValidacao(t *testing.T) {
	casos := []string{
		"/estoque/insumos",
		"/estoque/insumos?descricao=a",
		"/estoque/insumos?codigo=INS-1&quantidadeDesejada=0",
		"/estoque/insumos?codigo=INS-1&quantidadeDesejada=1.0001",
		"/estoque/insumos?codigo=INS-1&somenteDisponiveis=true",
		"/estoque/insumos?codigo=INS-1&tamanho=999",
		"/estoque/insumos?categoriaId=abc",
	}
	for _, alvo := range casos {
		t.Run(alvo, func(t *testing.T) {
			response := executarConsulta(t, alvo, consultarRepositorioFake{})
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, esperado 400: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestConsultarInsumoPorID(t *testing.T) {
	handler := NewConsultarInsumoPorIDHandler(insumo.NewConsultarInsumos(consultarRepositorioFake{itens: []insumoDomain.Insumo{insumoExemplo()}}))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/estoque/insumos/50000000-0000-0000-0000-000000000003?quantidadeDesejada=19", nil)
	request.SetPathValue("insumoId", "50000000-0000-0000-0000-000000000003")
	handler(response, request)

	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `"data"`) {
		t.Fatalf("detalhe incorreto: %d %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"quantidadeDisponivel":false`) {
		t.Fatalf("deveria indicar quantidade insuficiente: %s", response.Body.String())
	}
}

func TestConsultarInsumoPorIDErros(t *testing.T) {
	handler := NewConsultarInsumoPorIDHandler(insumo.NewConsultarInsumos(consultarRepositorioFake{}))
	casos := []struct {
		id     string
		status int
	}{
		{"abc", http.StatusBadRequest},
		{"00000000-0000-0000-0000-0000000000ff", http.StatusNotFound},
	}
	for _, caso := range casos {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/estoque/insumos/"+caso.id, nil)
		request.SetPathValue("insumoId", caso.id)
		handler(response, request)
		if response.Code != caso.status {
			t.Fatalf("status = %d, esperado %d", response.Code, caso.status)
		}
	}
}
