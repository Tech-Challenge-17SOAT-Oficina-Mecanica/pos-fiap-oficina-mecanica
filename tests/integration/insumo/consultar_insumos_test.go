package insumo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	segurancaDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	infra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/insumo"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/insumo"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestConsultarInsumos(t *testing.T) {
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(context.Background()); err != nil {
		t.Skip("banco indisponível")
	}

	jwt, err := seguranca.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{segurancaDominio.EscopoEstoqueLer})
	semEscopo, _ := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{segurancaDominio.EscopoEstoqueEscrever})

	useCase := application.NewConsultarInsumos(infra.NewPostgresRepository(db))
	listar := segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueLer, presentation.NewConsultarInsumosHandler(useCase))
	detalhar := segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueLer, presentation.NewConsultarInsumoPorIDHandler(useCase))

	t.Run("busca por codigo com disponibilidade", func(t *testing.T) {
		response := executarGET(listar, "/estoque/insumos?codigo=INS-000001&quantidadeDesejada=1", token, "")
		body := response.Body.String()
		if response.Code != http.StatusOK || !strings.Contains(body, `"codigo":"INS-000001"`) || !strings.Contains(body, `"quantidadeDisponivel":true`) {
			t.Fatalf("status %d: %s", response.Code, body)
		}
	})

	t.Run("somente disponiveis exclui insuficiente", func(t *testing.T) {
		response := executarGET(listar, "/estoque/insumos?descricao=Oleo&quantidadeDesejada=19&somenteDisponiveis=true", token, "")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"data":[]`) {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("inativo aparece somente quando solicitado", func(t *testing.T) {
		padrao := executarGET(listar, "/estoque/insumos?descricao=descontinuado", token, "")
		if padrao.Code != http.StatusOK || !strings.Contains(padrao.Body.String(), `"data":[]`) {
			t.Fatalf("inativo apareceu na listagem padrão: %d %s", padrao.Code, padrao.Body.String())
		}
		comInativo := executarGET(listar, "/estoque/insumos?descricao=descontinuado&incluirInativos=true", token, "")
		if comInativo.Code != http.StatusOK || !strings.Contains(comInativo.Body.String(), `"ativo":false`) {
			t.Fatalf("inativo não apareceu quando solicitado: %d %s", comInativo.Code, comInativo.Body.String())
		}
	})

	t.Run("detalhe direto devolve version", func(t *testing.T) {
		const id = "50000000-0000-0000-0000-000000000003"
		response := executarGET(detalhar, "/estoque/insumos/"+id+"?quantidadeDesejada=19", token, id)
		body := response.Body.String()
		if response.Code != http.StatusOK || strings.Contains(body, `"data"`) || !strings.Contains(body, `"version":1`) || !strings.Contains(body, `"quantidadeDisponivel":false`) {
			t.Fatalf("status %d: %s", response.Code, body)
		}
	})

	t.Run("contrato e autorizacao", func(t *testing.T) {
		casos := []struct {
			nome, url, token, id string
			handler              http.Handler
			status               int
		}{
			{"sem token", "/estoque/insumos?codigo=INS-000001", "", "", listar, http.StatusUnauthorized},
			{"sem escopo", "/estoque/insumos?codigo=INS-000001", semEscopo, "", listar, http.StatusForbidden},
			{"sem filtro", "/estoque/insumos", token, "", listar, http.StatusBadRequest},
			{"somente disponiveis sem quantidade", "/estoque/insumos?codigo=INS-000001&somenteDisponiveis=true", token, "", listar, http.StatusBadRequest},
			{"uuid invalido", "/estoque/insumos/abc", token, "abc", detalhar, http.StatusBadRequest},
			{"nao encontrado", "/estoque/insumos/00000000-0000-0000-0000-0000000000ff", token, "00000000-0000-0000-0000-0000000000ff", detalhar, http.StatusNotFound},
		}
		for _, caso := range casos {
			t.Run(caso.nome, func(t *testing.T) {
				response := executarGET(caso.handler, caso.url, caso.token, caso.id)
				if response.Code != caso.status {
					t.Fatalf("status %d: %s", response.Code, response.Body.String())
				}
			})
		}
	})
}

func executarGET(handler http.Handler, url, token, id string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, url, nil)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if id != "" {
		request.SetPathValue("insumoId", id)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
