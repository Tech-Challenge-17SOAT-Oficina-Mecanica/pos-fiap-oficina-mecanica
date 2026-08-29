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

func TestConsultarVeiculo(t *testing.T) {
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
	agora := time.Now().UnixNano()
	clienteID := fmt.Sprintf("60000000-0000-0000-0000-%012x", agora&0xffffffffffff)
	placaAtiva := fmt.Sprintf("X%c%c1A%02d", 'A'+agora%26, 'A'+agora/26%26, agora%100)
	placaInativa := fmt.Sprintf("Y%c%c1B%02d", 'A'+agora%26, 'A'+agora/26%26, agora%100)
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", time.Now().UnixNano()%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (cliente_id,placa,marca,modelo,ano,ativo) VALUES ($1,$2,'Toyota','Corolla',2020,TRUE),($1,$3,'Ford','Ka',2017,FALSE)", clienteID, placaAtiva, placaInativa); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE cliente_id=$1", clienteID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

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
	}{{"/veiculos?placa=" + placaAtiva + "&incluirInativos=false", allowed, 200}, {"/veiculos?placa=" + placaInativa + "&incluirInativos=false", allowed, 404}, {"/veiculos?placa=" + placaInativa + "&incluirInativos=true", allowed, 200}, {"/veiculos?placa=INVALIDA&incluirInativos=false", allowed, 400}, {"/veiculos", allowed, 400}, {"/veiculos?placa=" + placaAtiva + "&incluirInativos=x", allowed, 400}, {"/veiculos?placa=" + placaAtiva + "&incluirInativos=false", "", 401}, {"/veiculos?placa=" + placaAtiva + "&incluirInativos=false", denied, 403}} {
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
