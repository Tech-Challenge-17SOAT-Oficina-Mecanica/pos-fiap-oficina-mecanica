package integration_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
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
	mux := http.NewServeMux()
	mux.Handle("GET /veiculos", presentation.NewConsultaHandler(application.NewConsultar(infrastructure.NewPostgresRepository(db))))
	for _, test := range []struct {
		url    string
		status int
	}{{"/veiculos?placa=ABC1D23&incluirInativos=false", 200}, {"/veiculos?placa=GHI4J56&incluirInativos=false", 404}, {"/veiculos?placa=GHI4J56&incluirInativos=true", 200}, {"/veiculos?placa=INVALIDA&incluirInativos=false", 400}, {"/veiculos", 400}, {"/veiculos?placa=ABC1D23&incluirInativos=x", 400}} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", test.url, nil))
		if w.Code != test.status {
			t.Fatalf("%s: %d", test.url, w.Code)
		}
	}
}
