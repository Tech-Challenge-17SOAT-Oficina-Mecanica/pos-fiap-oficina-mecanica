package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestConsultarFornecedores(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"compras:ler"})
	if err != nil {
		t.Fatal(err)
	}
	semPermissao, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"mecanicos:escrever"})
	if err != nil {
		t.Fatal(err)
	}

	repository := infrastructure.NewPostgresRepository(db)
	mux := http.NewServeMux()
	mux.Handle("GET /fornecedores", segurancaPresentation.RequireScope(jwt, "compras:ler", presentation.NewListarHandler(application.NewConsultarFornecedores(repository))))
	mux.Handle("GET /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:ler", presentation.NewBuscarPorIDHandler(application.NewConsultarFornecedorPorID(repository))))

	tests := []struct {
		name   string
		path   string
		token  string
		status int
		body   string
	}{
		{name: "lista com envelope", path: "/fornecedores?documento=55666777000190", token: token, status: http.StatusOK, body: `"totalElementos":1`},
		{name: "consulta por id", path: "/fornecedores/60000000-0000-0000-0000-000000000001", token: token, status: http.StatusOK, body: `"version"`},
		{name: "inexistente", path: "/fornecedores/00000000-0000-0000-0000-000000000000", token: token, status: http.StatusNotFound},
		{name: "sem token", path: "/fornecedores", status: http.StatusUnauthorized},
		{name: "sem escopo", path: "/fornecedores", token: semPermissao, status: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body=%s", response.Body.String())
			}
		})
	}
}
