package servico_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	infra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/servico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestConsultarServicos(t *testing.T) {
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()

	jwt, err := seguranca.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"servicos:ler"})
	if err != nil {
		t.Fatal(err)
	}
	semEscopo, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"servicos:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	useCase := application.NewConsultar(infra.NewPostgresRepository(db))
	listar := segurancaPresentation.RequireScope(jwt, "servicos:ler", presentation.NewListarHandler(useCase))
	consultar := segurancaPresentation.RequireScope(jwt, "servicos:ler", presentation.NewConsultarHandler(useCase))

	t.Run("listagem padrão paginada oculta inativos", func(t *testing.T) {
		response := executarGET(listar, "/servicos?pagina=0&tamanho=2", token, "")
		if response.Code != http.StatusOK {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
		var output struct {
			Data []struct {
				ID, Codigo, Nome string
				Ativo            bool
			}
			Pagina, Tamanho, TotalElementos, TotalPaginas int
		}
		if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
			t.Fatal(err)
		}
		if output.Pagina != 0 || output.Tamanho != 2 || output.TotalElementos < 2 || output.TotalPaginas < 1 || len(output.Data) != 2 {
			t.Fatalf("envelope: %+v", output)
		}
		for _, servico := range output.Data {
			if !servico.Ativo {
				t.Fatalf("serviço inativo na listagem padrão: %+v", servico)
			}
		}
	})

	t.Run("filtro normalizado por nome", func(t *testing.T) {
		response := executarGET(listar, "/servicos?nome=%C3%93LEO", token, "")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Troca de oleo") {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("inclui inativos", func(t *testing.T) {
		response := executarGET(listar, "/servicos?nome=alinhamento&incluirInativos=true", token, "")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ativo":false`) {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
	})

	t.Run("detalhe direto com version e decimal exato", func(t *testing.T) {
		const id = "40000000-0000-0000-0000-000000000001"
		response := executarGET(consultar, "/servicos/"+id, token, id)
		body := response.Body.String()
		if response.Code != http.StatusOK || strings.Contains(body, `"data"`) || !strings.Contains(body, `"version":1`) || !strings.Contains(body, `"valor":150.00`) {
			t.Fatalf("status %d: %s", response.Code, body)
		}
	})

	t.Run("erros de contrato e autorização", func(t *testing.T) {
		cases := []struct {
			name, url, token, id string
			handler              http.Handler
			status               int
		}{
			{"sem autenticação", "/servicos", "", "", listar, http.StatusUnauthorized},
			{"sem permissão", "/servicos", semEscopo, "", listar, http.StatusForbidden},
			{"tamanho acima do teto", "/servicos?tamanho=51", token, "", listar, http.StatusBadRequest},
			{"uuid inválido", "/servicos/abc", token, "abc", consultar, http.StatusBadRequest},
			{"não encontrado", "/servicos/00000000-0000-0000-0000-0000000000ff", token, "00000000-0000-0000-0000-0000000000ff", consultar, http.StatusNotFound},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				response := executarGET(test.handler, test.url, test.token, test.id)
				if response.Code != test.status {
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
		request.SetPathValue("servicoId", id)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
