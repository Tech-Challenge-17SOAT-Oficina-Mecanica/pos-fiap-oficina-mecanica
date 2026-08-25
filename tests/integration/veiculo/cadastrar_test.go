package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
)

func TestCadastrarVeiculo(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL não configurada")
	}
	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	id := fmt.Sprintf("00000000-0000-0000-0000-%012x", time.Now().UnixNano()&0xffffffffffff)
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", id, fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE cliente_id=$1", id)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", id)
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
	handler := presentation.NewHandler(application.NewCadastrar(infrastructure.NewPostgresRepository(db)))
	mux.Handle("POST /clientes/{clienteId}/veiculos", securityPresentation.RequireScope(jwt, "veiculos:escrever", handler))
	placa := fmt.Sprintf("XYZ1A%02d", time.Now().UnixNano()%100)
	request := func(placa, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/clientes/"+id+"/veiculos", bytes.NewBufferString(`{"placa":"`+placa+`","marca":"Toyota","modelo":"Corolla","ano":2024}`))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	if w := request(placa, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("status sem token=%d", w.Code)
	}
	if w := request(placa, withoutScope); w.Code != http.StatusForbidden {
		t.Fatalf("status sem escopo=%d", w.Code)
	}
	if w := request(placa, authorized); w.Code != http.StatusCreated {
		t.Fatalf("status=%d: %s", w.Code, w.Body)
	}
	if w := request(placa, authorized); w.Code != http.StatusConflict {
		t.Fatalf("status duplicado=%d", w.Code)
	}
}
