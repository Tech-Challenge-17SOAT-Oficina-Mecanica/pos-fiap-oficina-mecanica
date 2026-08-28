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

type atualizarRepositorioFake struct {
	atualizada      pecaDomain.Peca
	erro            error
	versionRecebida int
	recebido        pecaDomain.Atualizacao
	usuarioID       string
}

func (fake *atualizarRepositorioFake) Atualizar(_ context.Context, _ string, version int, atualizacao pecaDomain.Atualizacao, usuarioID string) (pecaDomain.Peca, error) {
	fake.versionRecebida = version
	fake.recebido = atualizacao
	fake.usuarioID = usuarioID
	return fake.atualizada, fake.erro
}

const pecaID = "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4"

const corpoAtualizacao = `{
	"nome": "Pastilha de freio",
	"descricao": "Pastilha de freio dianteira cerâmica",
	"categoriaId": "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
	"fabricante": "Bosch",
	"precoVenda": 199.90,
	"estoqueMinimo": 6
}`

func atualizar(t *testing.T, ifMatch, corpo string, fake *atualizarRepositorioFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/estoque/pecas/"+pecaID, strings.NewReader(corpo))
	request.SetPathValue("pecaId", pecaID)
	if ifMatch != "" {
		request.Header.Set("If-Match", ifMatch)
	}
	NewAtualizarPecaHandler(peca.NewAtualizarPeca(fake)).ServeHTTP(response, request)
	return response
}

func fakeComSucesso() *atualizarRepositorioFake {
	atualizadoEm := time.Date(2026, 8, 26, 14, 30, 0, 0, time.UTC)
	usuario := "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
	preco := "199.90"
	return &atualizarRepositorioFake{atualizada: pecaDomain.Peca{
		ID: pecaID, Codigo: "PEC-000142", Nome: "Pastilha de freio",
		Descricao:   "Pastilha de freio dianteira cerâmica",
		CategoriaID: "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94", Categoria: "Freios",
		UnidadeMedida: "UN", PrecoVenda: &preco, EstoqueMinimo: 6,
		Ativo: true, Version: 8,
		DataAtualizacao: &atualizadoEm, UsuarioAtualizacao: &usuario,
	}}
}

func TestAtualizarPecaRetorna200(t *testing.T) {
	fake := fakeComSucesso()
	response := atualizar(t, "7", corpoAtualizacao, fake)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200. corpo=%s", response.Code, response.Body.String())
	}
	if fake.versionRecebida != 7 {
		t.Fatalf("version repassada = %d, esperado 7", fake.versionRecebida)
	}

	var corpo map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}
	for _, campo := range []string{"atualizadoEm", "atualizadoPor"} {
		if _, presente := corpo[campo]; !presente {
			t.Fatalf("%q deveria estar na resposta da atualização", campo)
		}
	}
	if corpo["version"] != float64(8) {
		t.Fatalf("version = %v, esperado 8", corpo["version"])
	}
}

// O bug que existia nos outros módulos: aspas são a forma correta de ETag por HTTP.
func TestAtualizarPecaAceitaIfMatchComAspas(t *testing.T) {
	for _, header := range []string{"7", `"7"`, ` "7" `} {
		fake := fakeComSucesso()
		response := atualizar(t, header, corpoAtualizacao, fake)
		if response.Code != http.StatusOK {
			t.Fatalf("If-Match %q: status = %d, esperado 200", header, response.Code)
		}
		if fake.versionRecebida != 7 {
			t.Fatalf("If-Match %q: version = %d, esperado 7", header, fake.versionRecebida)
		}
	}
}

func TestAtualizarPecaErros(t *testing.T) {
	casos := []struct {
		nome    string
		ifMatch string
		corpo   string
		erro    error
		status  int
		campo   string
	}{
		{"sem If-Match", "", corpoAtualizacao, nil, http.StatusPreconditionRequired, "If-Match"},
		{"If-Match invalido", "abc", corpoAtualizacao, nil, http.StatusBadRequest, "If-Match"},
		{"json quebrado", "7", `{`, nil, http.StatusBadRequest, ""},
		{"ativo no corpo", "7", `{"nome":"Peca","descricao":"Valida","categoriaId":"7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94","precoVenda":10,"ativo":false}`, nil, http.StatusBadRequest, "ativo"},
		{"preco ausente", "7", `{"nome":"Peca","descricao":"Valida","categoriaId":"7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94"}`, nil, http.StatusBadRequest, "precoVenda"},
		{"version divergente", "7", corpoAtualizacao, peca.ErrVersaoDivergente, http.StatusPreconditionFailed, "If-Match"},
		{"nao encontrada", "7", corpoAtualizacao, peca.ErrNaoEncontrada, http.StatusNotFound, ""},
		{"categoria invalida", "7", corpoAtualizacao, peca.ErrCategoriaInvalida, http.StatusBadRequest, "categoriaId"},
		{"descricao duplicada", "7", corpoAtualizacao, peca.ErrDescricaoDuplicada, http.StatusConflict, "descricao"},
		{"falha inesperada", "7", corpoAtualizacao, context.DeadlineExceeded, http.StatusInternalServerError, ""},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			response := atualizar(t, caso.ifMatch, caso.corpo, &atualizarRepositorioFake{erro: caso.erro})
			if response.Code != caso.status {
				t.Fatalf("status = %d, esperado %d. corpo=%s", response.Code, caso.status, response.Body.String())
			}
			if tipo := response.Header().Get("Content-Type"); tipo != "application/problem+json" {
				t.Fatalf("Content-Type = %q", tipo)
			}
			if caso.campo != "" && !strings.Contains(response.Body.String(), `"campo":"`+caso.campo+`"`) {
				t.Fatalf("erro deveria apontar %q: %s", caso.campo, response.Body.String())
			}
		})
	}
}
