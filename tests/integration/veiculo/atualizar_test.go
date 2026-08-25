package integration_test

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
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
	now := time.Now().UnixNano()
	placa := fmt.Sprintf("X%c%c1A%02d", 'A'+now%26, 'A'+now/26%26, now%100)
	if _, err = db.Exec("INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Toyota','Corolla',2020)", id, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM veiculo WHERE id=$1", id)
		_, _ = db.Exec("DELETE FROM cliente WHERE id=$1", clienteID)
	})
	jwt, err := securityInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := jwt.Gerar("usuario", []string{"veiculos:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	withoutScope, err := jwt.Gerar("usuario", nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	handler := presentation.NewAtualizarHandler(application.NewAtualizar(infrastructure.NewPostgresRepository(db)))
	mux.Handle("PUT /veiculos/{veiculoId}", securityPresentation.RequireScope(jwt, "veiculos:escrever", handler))
	req := func(id, match, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("PUT", "/veiculos/"+id, bytes.NewBufferString(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		if match != "" {
			r.Header.Set("If-Match", match)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	body := `{"placa":"` + placa + `","marca":"Toyota","modelo":"Corolla XEi","ano":2021}`
	if w := req(id, "", body, ""); w.Code != 401 {
		t.Fatalf("sem token: %d", w.Code)
	}
	if w := req(id, "", body, withoutScope); w.Code != 403 {
		t.Fatalf("sem escopo: %d", w.Code)
	}
	if w := req(id, "", body, authorized); w.Code != 428 {
		t.Fatalf("sem If-Match: %d", w.Code)
	}
	if w := req(id, "999", body, authorized); w.Code != 412 {
		t.Fatalf("versão: %d", w.Code)
	}
	if w := req("30000000-0000-0000-0000-000000000099", "1", body, authorized); w.Code != 404 {
		t.Fatalf("inexistente: %d", w.Code)
	}
	if w := req(id, "1", `{"placa":"DEF1234","marca":"Toyota","modelo":"Corolla","ano":2021}`, authorized); w.Code != 409 {
		t.Fatalf("duplicada: %d", w.Code)
	}
	if w := req(id, "1", body, authorized); w.Code != 200 {
		t.Fatalf("sucesso: %d", w.Code)
	}
}
