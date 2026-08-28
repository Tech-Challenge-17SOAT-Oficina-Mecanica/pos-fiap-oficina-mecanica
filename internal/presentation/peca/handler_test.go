package peca

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type repositorioFake struct {
	pecas []pecaDomain.Peca
	total int
	erro  error
}

func (fake repositorioFake) BuscarPorFiltro(context.Context, peca.Filtros, int, int) ([]pecaDomain.Peca, int, error) {
	return fake.pecas, fake.total, fake.erro
}

func (fake repositorioFake) BuscarPorID(context.Context, string) (pecaDomain.Peca, error) {
	if len(fake.pecas) == 0 {
		return pecaDomain.Peca{}, peca.ErrNaoEncontrada
	}
	return fake.pecas[0], fake.erro
}

func pecaExemplo() pecaDomain.Peca {
	preco := "189.90"
	fabricante := "Bosch"
	fornecedorID := "60000000-0000-0000-0000-000000000001"
	return pecaDomain.Peca{
		ID: "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", Codigo: "PEC-000142",
		Nome: "Pastilha de freio", Descricao: "Pastilha de freio dianteira",
		CategoriaID: "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94", Categoria: "Freios",
		FornecedorID: &fornecedorID, Fabricante: &fabricante, UnidadeMedida: "UN", PrecoVenda: &preco,
		SaldoFisico: 6, SaldoReservado: 2, EstoqueMinimo: 4, Ativo: true, Version: 3,
	}
}

func executar(t *testing.T, alvo string, fake repositorioFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler := NewConsultarPecasHandler(peca.NewConsultarPecas(fake))
	handler(response, httptest.NewRequest("GET", alvo, nil))
	return response
}

func TestConsultarPecasRetornaEnvelopeCompleto(t *testing.T) {
	response := executar(t, "/estoque/pecas?codigo=PEC-000142&quantidadeDesejada=3",
		repositorioFake{pecas: []pecaDomain.Peca{pecaExemplo()}, total: 1})

	if response.Code != 200 {
		t.Fatalf("status = %d, corpo %s", response.Code, response.Body)
	}

	var corpo struct {
		Data []struct {
			Codigo               string      `json:"codigo"`
			Tipo                 string      `json:"tipo"`
			FornecedorID         string      `json:"fornecedorId"`
			PrecoVenda           json.Number `json:"precoVenda"`
			SaldoDisponivel      int64       `json:"saldoDisponivel"`
			QuantidadeDesejada   *int64      `json:"quantidadeDesejada"`
			QuantidadeDisponivel *bool       `json:"quantidadeDisponivel"`
			Disponivel           bool        `json:"disponivel"`
			AbaixoDoMinimo       bool        `json:"abaixoDoMinimo"`
			Version              int         `json:"version"`
		} `json:"data"`
		Pagina         int `json:"pagina"`
		Tamanho        int `json:"tamanho"`
		TotalElementos int `json:"totalElementos"`
		TotalPaginas   int `json:"totalPaginas"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}

	if len(corpo.Data) != 1 {
		t.Fatalf("esperado 1 item, recebido %d", len(corpo.Data))
	}
	item := corpo.Data[0]
	if item.Tipo != "PECA" || item.Codigo != "PEC-000142" {
		t.Fatalf("identificacao errada: %+v", item)
	}
	if item.FornecedorID != "60000000-0000-0000-0000-000000000001" {
		t.Fatalf("fornecedorId = %q", item.FornecedorID)
	}
	if item.PrecoVenda.String() != "189.90" {
		t.Fatalf("precoVenda = %s, esperado 189.90 sem arredondamento", item.PrecoVenda)
	}
	if item.SaldoDisponivel != 4 || !item.Disponivel || item.AbaixoDoMinimo {
		t.Fatalf("saldos derivados errados: %+v", item)
	}
	if item.QuantidadeDesejada == nil || *item.QuantidadeDesejada != 3 {
		t.Fatalf("quantidadeDesejada ausente: %+v", item)
	}
	if item.QuantidadeDisponivel == nil || !*item.QuantidadeDisponivel {
		t.Fatalf("quantidadeDisponivel deveria ser true: %+v", item)
	}
	if corpo.Tamanho != 20 || corpo.TotalElementos != 1 || corpo.TotalPaginas != 1 {
		t.Fatalf("envelope errado: %+v", corpo)
	}
}

func TestConsultarPecasOmiteQuantidadeQuandoNaoInformada(t *testing.T) {
	response := executar(t, "/estoque/pecas?codigo=PEC-000142",
		repositorioFake{pecas: []pecaDomain.Peca{pecaExemplo()}, total: 1})

	if response.Body.String() == "" {
		t.Fatal("corpo vazio")
	}
	if strings.Contains(response.Body.String(), "quantidadeDesejada") {
		t.Fatalf("quantidadeDesejada nao deveria aparecer: %s", response.Body)
	}
}

func TestConsultarPecasListaVaziaRetorna200(t *testing.T) {
	response := executar(t, "/estoque/pecas?codigo=PEC-999999", repositorioFake{})

	if response.Code != 200 {
		t.Fatalf("status = %d, esperado 200 com data vazio", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"data":[]`) {
		t.Fatalf("data deveria ser [], recebido %s", response.Body)
	}
}

func TestConsultarPecasErrosDeValidacao(t *testing.T) {
	casos := []struct {
		nome   string
		alvo   string
		status int
	}{
		{"sem filtro de busca", "/estoque/pecas", 400},
		{"descricao curta", "/estoque/pecas?descricao=a", 400},
		{"quantidade zero", "/estoque/pecas?codigo=PEC-1&quantidadeDesejada=0", 400},
		{"quantidade nao numerica", "/estoque/pecas?codigo=PEC-1&quantidadeDesejada=abc", 400},
		{"tamanho acima do teto", "/estoque/pecas?codigo=PEC-1&tamanho=999", 400},
		{"pagina negativa", "/estoque/pecas?codigo=PEC-1&pagina=-1", 400},
		{"categoria invalida", "/estoque/pecas?codigo=PEC-1&categoriaId=abc", 400},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			response := executar(t, caso.alvo, repositorioFake{})
			if response.Code != caso.status {
				t.Fatalf("status = %d, esperado %d, corpo %s", response.Code, caso.status, response.Body)
			}
			if response.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestConsultarPecaPorID(t *testing.T) {
	handler := NewConsultarPecaPorIDHandler(peca.NewConsultarPecas(repositorioFake{pecas: []pecaDomain.Peca{pecaExemplo()}}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/estoque/pecas/3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", nil)
	request.SetPathValue("pecaId", "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4")
	handler(response, request)

	if response.Code != 200 {
		t.Fatalf("status = %d, corpo %s", response.Code, response.Body)
	}
	if strings.Contains(response.Body.String(), `"data"`) {
		t.Fatalf("recurso unico nao deve usar envelope: %s", response.Body)
	}
	if !strings.Contains(response.Body.String(), `"version":3`) {
		t.Fatalf("detalhe deve expor version: %s", response.Body)
	}
}

func TestConsultarPecaPorIDInexistenteRetorna404(t *testing.T) {
	handler := NewConsultarPecaPorIDHandler(peca.NewConsultarPecas(repositorioFake{}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/estoque/pecas/3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", nil)
	request.SetPathValue("pecaId", "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4")
	handler(response, request)

	if response.Code != 404 {
		t.Fatalf("status = %d, esperado 404", response.Code)
	}
}

func TestConsultarPecaPorIDIdentificadorInvalido(t *testing.T) {
	handler := NewConsultarPecaPorIDHandler(peca.NewConsultarPecas(repositorioFake{}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/estoque/pecas/abc", nil)
	request.SetPathValue("pecaId", "abc")
	handler(response, request)

	if response.Code != 400 {
		t.Fatalf("status = %d, esperado 400", response.Code)
	}
}
