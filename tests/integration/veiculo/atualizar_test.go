package integration_test

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAtualizarVeiculo(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL não configurada")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	id := fmt.Sprintf("40000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffffff)
	clienteID := fmt.Sprintf("50000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffffff)
	if _, err = db.Exec("INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)); err != nil { t.Fatal(err) }
	if _, err = db.Exec("INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,'XYZ1A99','Toyota','Corolla',2020)", id, clienteID); err != nil { t.Fatal(err) }
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM veiculo WHERE id=$1", id); _, _ = db.Exec("DELETE FROM cliente WHERE id=$1", clienteID) })
	mux := http.NewServeMux()
	mux.Handle("PUT /veiculos/{veiculoId}", presentation.NewAtualizarHandler(application.NewAtualizar(infrastructure.NewPostgresRepository(db))))
	req := func(id, match, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("PUT", "/veiculos/"+id, bytes.NewBufferString(body))
		if match != "" {
			r.Header.Set("If-Match", match)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	body := `{"placa":"XYZ1A99","marca":"Toyota","modelo":"Corolla XEi","ano":2021}`
	if w := req(id, "", body); w.Code != 428 {
		t.Fatalf("sem If-Match: %d", w.Code)
	}
	if w := req(id, "999", body); w.Code != 412 {
		t.Fatalf("versão: %d", w.Code)
	}
	if w := req("30000000-0000-0000-0000-000000000099", "1", body); w.Code != 404 {
		t.Fatalf("inexistente: %d", w.Code)
	}
	if w := req(id, "1", `{"placa":"DEF1234","marca":"Toyota","modelo":"Corolla","ano":2021}`); w.Code != 409 {
		t.Fatalf("duplicada: %d", w.Code)
	}
	if w := req(id, "1", body); w.Code != 200 {
		t.Fatalf("sucesso: %d", w.Code)
	}
}
