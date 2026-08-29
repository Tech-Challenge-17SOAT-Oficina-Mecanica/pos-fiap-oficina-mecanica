package insumo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	insumoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

func textoInsumo(valor string) *string { return &valor }

type cadastrarRepositorioFake struct {
	cadastrado insumoDomain.Insumo
	erro       error
	recebido   insumoDomain.Cadastro
}

func (fake *cadastrarRepositorioFake) Cadastrar(_ context.Context, cadastro insumoDomain.Cadastro) (insumoDomain.Insumo, error) {
	fake.recebido = cadastro
	return fake.cadastrado, fake.erro
}

func cadastrar(t *testing.T, corpo string, fake *cadastrarRepositorioFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/estoque/insumos", strings.NewReader(corpo))
	NewCadastrarInsumoHandler(insumo.NewCadastrarInsumo(fake)).ServeHTTP(response, request)
	return response
}

const corpoValido = `{
	"nome": "Óleo 5W30",
	"descricao": "Óleo sintético 5W30 API SN",
	"categoriaId": "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724",
	"fornecedorId": "60000000-0000-0000-0000-000000000001",
	"unidadeMedida": "L",
	"custoUnitario": 45.0,
	"estoqueMinimo": 20.5
}`

func TestCadastrarInsumoRetorna201(t *testing.T) {
	criadoEm := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	custo := "45.00"
	fake := &cadastrarRepositorioFake{cadastrado: insumoDomain.Insumo{
		ID:             "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
		Codigo:         "INS-000004",
		Nome:           "Óleo 5W30",
		Descricao:      "Óleo sintético 5W30 API SN",
		CategoriaID:    "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724",
		FornecedorID:   textoInsumo("60000000-0000-0000-0000-000000000001"),
		Categoria:      "Lubrificantes",
		UnidadeMedida:  "L",
		CustoUnitario:  &custo,
		SaldoFisico:    "0.000",
		SaldoReservado: "0.000",
		EstoqueMinimo:  "20.500",
		Ativo:          true,
		Version:        1,
		DataCriacao:    &criadoEm,
	}}

	response := cadastrar(t, corpoValido, fake)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado 201. corpo=%s", response.Code, response.Body.String())
	}

	var corpo map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}
	if corpo["codigo"] != "INS-000004" {
		t.Fatalf("codigo = %v", corpo["codigo"])
	}
	if corpo["tipo"] != "INSUMO" {
		t.Fatalf("tipo = %v", corpo["tipo"])
	}
	if corpo["unidadeMedida"] != "L" {
		t.Fatalf("unidadeMedida = %v", corpo["unidadeMedida"])
	}
	if corpo["fornecedorId"] != "60000000-0000-0000-0000-000000000001" {
		t.Fatalf("fornecedorId = %v", corpo["fornecedorId"])
	}

	// O decimal precisa sair como número, sem virar string nem perder a fração.
	if !strings.Contains(response.Body.String(), `"estoqueMinimo":20.500`) {
		t.Fatalf("estoqueMinimo deveria sair como decimal: %s", response.Body.String())
	}

	if fake.recebido.DescricaoNormalizada != "oleo sintetico 5w30 api sn" {
		t.Fatalf("descricaoNormalizada = %q", fake.recebido.DescricaoNormalizada)
	}
	if fake.recebido.EstoqueMinimo != "20.5" {
		t.Fatalf("estoqueMinimo repassado = %q; a fração deveria chegar intacta", fake.recebido.EstoqueMinimo)
	}
	if fake.recebido.FornecedorID == nil || *fake.recebido.FornecedorID != "60000000-0000-0000-0000-000000000001" {
		t.Fatalf("fornecedorID repassado = %v", fake.recebido.FornecedorID)
	}
}

func TestCadastrarInsumoErros(t *testing.T) {
	casos := []struct {
		nome   string
		corpo  string
		erro   error
		status int
	}{
		{"json quebrado", `{`, nil, http.StatusBadRequest},
		{"unidade fora do enum", `{"nome":"Item","descricao":"Valida","categoriaId":"e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724","unidadeMedida":"CX","custoUnitario":1}`, nil, http.StatusBadRequest},
		{"custo ausente", `{"nome":"Item","descricao":"Valida","categoriaId":"e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724","unidadeMedida":"L"}`, nil, http.StatusBadRequest},
		{"categoria nao uuid", `{"nome":"Item","descricao":"Valida","categoriaId":"abc","unidadeMedida":"L","custoUnitario":1}`, nil, http.StatusBadRequest},
		{"categoria inativa", corpoValido, insumo.ErrCategoriaInvalida, http.StatusBadRequest},
		{"fornecedor invalido", corpoValido, insumo.ErrFornecedorInvalido, http.StatusBadRequest},
		{"descricao duplicada", corpoValido, insumo.ErrDescricaoDuplicada, http.StatusConflict},
		{"falha inesperada", corpoValido, context.DeadlineExceeded, http.StatusInternalServerError},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			response := cadastrar(t, caso.corpo, &cadastrarRepositorioFake{erro: caso.erro})
			if response.Code != caso.status {
				t.Fatalf("status = %d, esperado %d. corpo=%s", response.Code, caso.status, response.Body.String())
			}
			if tipo := response.Header().Get("Content-Type"); tipo != "application/problem+json" {
				t.Fatalf("Content-Type = %q", tipo)
			}
		})
	}
}

// Regressão do apontamento de revisão: o erro precisa apontar o parâmetro culpado.
func TestCadastroInvalidoApontaCampo(t *testing.T) {
	casos := []struct {
		nome  string
		corpo string
		campo string
	}{
		{"nome", `{"nome":"","descricao":"Valida","categoriaId":"e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724","unidadeMedida":"L","custoUnitario":1}`, "nome"},
		{"unidadeMedida", `{"nome":"Item","descricao":"Valida","categoriaId":"e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724","unidadeMedida":"CX","custoUnitario":1}`, "unidadeMedida"},
		{"custoUnitario", `{"nome":"Item","descricao":"Valida","categoriaId":"e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724","unidadeMedida":"L"}`, "custoUnitario"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			response := cadastrar(t, caso.corpo, &cadastrarRepositorioFake{})
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", response.Code)
			}
			if !strings.Contains(response.Body.String(), `"campo":"`+caso.campo+`"`) {
				t.Fatalf("erro deveria apontar %q: %s", caso.campo, response.Body.String())
			}
		})
	}
}
