package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

func TestConsultarOrcamento(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	jwt, err := securityInfrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, _ := jwt.Gerar("mecanico", []string{"os:ler"})
	cliente, _ := jwt.GerarCliente("20000000-0000-0000-0000-000000000001", "70000000-0000-0000-0000-000000000001")
	clienteInvalido, _ := jwt.GerarCliente("20000000-0000-0000-0000-000000000002", "70000000-0000-0000-0000-000000000001")
	mux := http.NewServeMux()
	mux.Handle("GET /ordens-servico/{osId}/orcamento", securityPresentation.RequireAnyScope(jwt, []string{"os:ler", "orcamentos:ler"}, presentation.NewConsultarHandler(application.NewConsultar(infrastructure.NewPostgresRepository(db)))))
	request := func(token, osID string) *httptest.ResponseRecorder {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ordens-servico/"+osID+"/orcamento", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		mux.ServeHTTP(writer, req)
		return writer
	}
	if writer := request("", "70000000-0000-0000-0000-000000000001"); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", writer.Code)
	}
	if writer := request(clienteInvalido, "70000000-0000-0000-0000-000000000001"); writer.Code != http.StatusForbidden {
		t.Fatalf("cliente invalido=%d", writer.Code)
	}
	if writer := request(cliente, "70000000-0000-0000-0000-000000000001"); writer.Code != http.StatusOK || !strings.Contains(writer.Body.String(), `"orcamentoId"`) {
		t.Fatalf("cliente=%d body=%s", writer.Code, writer.Body.String())
	}
	if writer := request(mecanico, "70000000-0000-0000-0000-000000000001"); writer.Code != http.StatusOK {
		t.Fatalf("mecanico=%d body=%s", writer.Code, writer.Body.String())
	}
	if writer := request(mecanico, "70000000-0000-0000-0000-000000000099"); writer.Code != http.StatusNotFound {
		t.Fatalf("ausente=%d", writer.Code)
	}
}
