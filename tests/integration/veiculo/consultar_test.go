package integration_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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
	sufixo := time.Now().UnixNano() % 100
	clienteID := fmt.Sprintf("60000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffffff)
	placaAtiva := fmt.Sprintf("XAA1A%02d", sufixo)
	placaInativa := fmt.Sprintf("XBB1B%02d", sufixo)
	if _, err = db.Exec("INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)); err != nil { t.Fatal(err) }
	if _, err = db.Exec("INSERT INTO veiculo (cliente_id,placa,marca,modelo,ano,ativo) VALUES ($1,$2,'Toyota','Corolla',2020,TRUE),($1,$3,'Ford','Ka',2017,FALSE)", clienteID, placaAtiva, placaInativa); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM veiculo WHERE cliente_id=$1", clienteID); _, _ = db.Exec("DELETE FROM cliente WHERE id=$1", clienteID) })
	mux := http.NewServeMux()
	mux.Handle("GET /veiculos", presentation.NewConsultaHandler(application.NewConsultar(infrastructure.NewPostgresRepository(db))))
	for _, test := range []struct {
		url    string
		status int
	}{{"/veiculos?placa="+placaAtiva+"&incluirInativos=false", 200}, {"/veiculos?placa="+placaInativa+"&incluirInativos=false", 404}, {"/veiculos?placa="+placaInativa+"&incluirInativos=true", 200}, {"/veiculos?placa=INVALIDA&incluirInativos=false", 400}, {"/veiculos", 400}, {"/veiculos?placa="+placaAtiva+"&incluirInativos=x", 400}} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", test.url, nil))
		if w.Code != test.status {
			t.Fatalf("%s: %d", test.url, w.Code)
		}
	}
}
