package peca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type cadastrarRepositorioFake struct {
	cadastrada pecaDomain.Peca
	erro       error
	recebido   pecaDomain.Cadastro
}

func (fake *cadastrarRepositorioFake) Cadastrar(_ context.Context, cadastro pecaDomain.Cadastro) (pecaDomain.Peca, error) {
	fake.recebido = cadastro
	return fake.cadastrada, fake.erro
}

func cadastrar(t *testing.T, corpo string, fake *cadastrarRepositorioFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/estoque/pecas", strings.NewReader(corpo))
	NewCadastrarPecaHandler(peca.NewCadastrarPeca(fake)).ServeHTTP(response, request)
	return response
}

const corpoValido = `{
	"nome": "Pastilha de freio",
	"descricao": "Pastilha de freio dianteira",
	"categoriaId": "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
	"fabricante": "Fabricante X",
	"precoVenda": 180.0,
	"estoqueMinimo": 4
}`

func TestCadastrarPecaRetorna201(t *testing.T) {
	criadoEm := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	preco := "180.00"
	fake := &cadastrarRepositorioFake{cadastrada: pecaDomain.Peca{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		Codigo:        "PEC-000003",
		Nome:          "Pastilha de freio",
		Descricao:     "Pastilha de freio dianteira",
		CategoriaID:   "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
		Categoria:     "Freios",
		UnidadeMedida: "UN",
		PrecoVenda:    &preco,
		EstoqueMinimo: 4,
		Ativo:         true,
		Version:       1,
		DataCriacao:   &criadoEm,
	}}

	response := cadastrar(t, corpoValido, fake)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado 201. corpo=%s", response.Code, response.Body.String())
	}

	var corpo map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}
	if corpo["codigo"] != "PEC-000003" {
		t.Fatalf("codigo = %v", corpo["codigo"])
	}
	if corpo["tipo"] != "PECA" {
		t.Fatalf("tipo = %v", corpo["tipo"])
	}
	if corpo["ativo"] != true {
		t.Fatalf("ativo = %v", corpo["ativo"])
	}
	if _, presente := corpo["dataCriacao"]; !presente {
		t.Fatal("dataCriacao deveria estar presente no cadastro")
	}

	// A normalização precisa chegar ao repositório, é ela que alimenta o índice único.
	if fake.recebido.DescricaoNormalizada != "pastilha de freio dianteira" {
		t.Fatalf("descricaoNormalizada = %q", fake.recebido.DescricaoNormalizada)
	}
	if fake.recebido.UnidadeMedida != "UN" {
		t.Fatalf("unidadeMedida = %q, esperado UN", fake.recebido.UnidadeMedida)
	}
}

func TestCadastrarPecaErros(t *testing.T) {
	casos := []struct {
		nome   string
		corpo  string
		erro   error
		status int
	}{
		{"json quebrado", `{`, nil, http.StatusBadRequest},
		{"nome vazio", `{"nome":"","descricao":"Valida","categoriaId":"7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94"}`, nil, http.StatusBadRequest},
		{"categoria nao uuid", `{"nome":"Peca","descricao":"Valida","categoriaId":"abc"}`, nil, http.StatusBadRequest},
		{"categoria inativa", corpoValido, peca.ErrCategoriaInvalida, http.StatusBadRequest},
		{"descricao duplicada", corpoValido, peca.ErrDescricaoDuplicada, http.StatusConflict},
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

func TestConsultarNaoDevolveDataCriacao(t *testing.T) {
	resposta := montarResponse(pecaDomain.Peca{ID: "id-1", Codigo: "PEC-000001"}, nil)
	corpo, err := json.Marshal(resposta)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(corpo), "dataCriacao") {
		t.Fatalf("dataCriacao vazou para a consulta: %s", corpo)
	}
}
