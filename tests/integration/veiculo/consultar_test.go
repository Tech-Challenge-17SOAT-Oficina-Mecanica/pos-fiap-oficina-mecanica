package integration_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
)

func TestConsultarVeiculo(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL não configurada")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	jwt, err := securityInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	allowed, err := jwt.Gerar("usuario", []string{"veiculos:ler"})
	if err != nil {
		t.Fatal(err)
	}
	denied, err := jwt.Gerar("usuario", nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	handler := presentation.NewConsultaHandler(application.NewConsultar(infrastructure.NewPostgresRepository(db)))
	mux.Handle("GET /veiculos", securityPresentation.RequireScope(jwt, "veiculos:ler", handler))
	for _, test := range []struct {
		url    string
		token  string
		status int
	}{{"/veiculos?placa=ABC1D23&incluirInativos=false", allowed, 200}, {"/veiculos?placa=GHI4J56&incluirInativos=false", allowed, 404}, {"/veiculos?placa=GHI4J56&incluirInativos=true", allowed, 200}, {"/veiculos?placa=INVALIDA&incluirInativos=false", allowed, 400}, {"/veiculos", allowed, 400}, {"/veiculos?placa=ABC1D23&incluirInativos=x", allowed, 400}, {"/veiculos?placa=ABC1D23&incluirInativos=false", "", 401}, {"/veiculos?placa=ABC1D23&incluirInativos=false", denied, 403}} {
		w := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, test.url, nil)
		if test.token != "" {
			request.Header.Set("Authorization", "Bearer "+test.token)
		}
		mux.ServeHTTP(w, request)
		if w.Code != test.status {
			t.Fatalf("%s: %d", test.url, w.Code)
		}
	}
}
