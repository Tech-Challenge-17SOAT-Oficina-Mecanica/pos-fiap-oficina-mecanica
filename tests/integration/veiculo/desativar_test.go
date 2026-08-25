package integration_test

import (
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

func TestDesativarVeiculo(t *testing.T) {
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
	now := time.Now().UnixNano()
	clienteID := fmt.Sprintf("90000000-0000-0000-0000-%012x", now&0xffffffffffff)
	veiculoID := fmt.Sprintf("91000000-0000-0000-0000-%012x", now&0xffffffffffff)
	placa := fmt.Sprintf("Z%c%c1A%02d", 'A'+now%26, 'A'+now/26%26, now%100)
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", now%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Toyota','Corolla',2020)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})
	jwt, _ := securityInfrastructure.NewJWT("segredo")
	allowed, _ := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"veiculos:escrever"})
	denied, _ := jwt.Gerar("90000000-0000-0000-0000-000000000001", nil)
	repo := infrastructure.NewPostgresRepository(db)
	mux := http.NewServeMux()
	mux.Handle("DELETE /veiculos/{veiculoId}", securityPresentation.RequireScope(jwt, "veiculos:escrever", presentation.NewInativarHandler(application.NewInativar(repo))))
	mux.Handle("POST /veiculos/{veiculoId}/reativacao", securityPresentation.RequireScope(jwt, "veiculos:escrever", presentation.NewReativarHandler(application.NewReativar(repo))))
	call := func(method, path, token string) int {
		r := httptest.NewRequest(method, path, nil)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w.Code
	}
	if got := call(http.MethodDelete, "/veiculos/"+veiculoID+"?motivo=teste", ""); got != 401 {
		t.Fatal(got)
	}
	if got := call(http.MethodDelete, "/veiculos/"+veiculoID, denied); got != 403 {
		t.Fatal(got)
	}
	if got := call(http.MethodDelete, "/veiculos/"+veiculoID+"?motivo=teste", allowed); got != 200 {
		t.Fatal(got)
	}
	if got := call(http.MethodDelete, "/veiculos/"+veiculoID, allowed); got != 204 {
		t.Fatal(got)
	}
	if got := call(http.MethodPost, "/veiculos/"+veiculoID+"/reativacao", allowed); got != 200 {
		t.Fatal(got)
	}
}
