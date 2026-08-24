package integration_test

import (
	"bytes"
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

func TestCadastrarVeiculo(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL não configurada")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	id := fmt.Sprintf("00000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffffff)
	if _, err = db.Exec("INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", id, fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM veiculo WHERE cliente_id=$1", id)
		_, _ = db.Exec("DELETE FROM cliente WHERE id=$1", id)
	})

	mux := http.NewServeMux()
	mux.Handle("POST /clientes/{clienteId}/veiculos", presentation.NewHandler(application.NewCadastrar(infrastructure.NewPostgresRepository(db))))
	placa := fmt.Sprintf("XYZ1A%02d", time.Now().UnixNano()%100)
	request := func(placa string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/clientes/"+id+"/veiculos", bytes.NewBufferString(`{"placa":"`+placa+`","marca":"Toyota","modelo":"Corolla","ano":2024}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	if w := request(placa); w.Code != http.StatusCreated {
		t.Fatalf("status=%d: %s", w.Code, w.Body)
	}
	if w := request(placa); w.Code != http.StatusConflict {
		t.Fatalf("status duplicado=%d", w.Code)
	}
}
