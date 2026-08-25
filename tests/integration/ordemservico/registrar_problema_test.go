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
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

func TestRegistrarProblemaEncontrado(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	suffix := time.Now().UnixNano() & 0xffffffffffff
	clienteID := fmt.Sprintf("81000000-0000-0000-0000-%012x", suffix)
	veiculoID := fmt.Sprintf("82000000-0000-0000-0000-%012x", suffix)
	osDiagnosticoID := fmt.Sprintf("83000000-0000-0000-0000-%012x", suffix)
	osFechadaID := fmt.Sprintf("84000000-0000-0000-0000-%012x", suffix)
	placa := fmt.Sprintf("TST1A%02d", suffix%100)
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES ($1,$2,$3,$4,'EM_DIAGNOSTICO'),($5,$2,$3,$4,'ENTREGUE')", osDiagnosticoID, clienteID, veiculoID, placa, osFechadaID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM problema_ordem_servico WHERE ordem_servico_id IN ($1,$2)", osDiagnosticoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE ordem_servico_id IN ($1,$2)", osDiagnosticoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2)", osDiagnosticoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

	jwt, err := securityInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("POST /ordens-servico/{osId}/problemas", securityPresentation.RequireScope(jwt, "os:escrever", presentation.NewRegistrarProblemaHandler(application.NewRegistrarProblema(infrastructure.NewPostgresRepository(db)))))
	request := func(osID, body string) *httptest.ResponseRecorder {
		writer := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/problemas", bytes.NewBufferString(body))
		httpRequest.Header.Set("Authorization", "Bearer "+token)
		mux.ServeHTTP(writer, httpRequest)
		return writer
	}
	if writer := request(osDiagnosticoID, `{"descricao":"Vazamento de oleo","observacoes":"na junta"}`); writer.Code != http.StatusCreated {
		t.Fatalf("cadastro: %d %s", writer.Code, writer.Body.String())
	}
	if writer := request(osDiagnosticoID, `{"descricao":"Filtro saturado"}`); writer.Code != http.StatusCreated {
		t.Fatalf("reuso do orcamento: %d %s", writer.Code, writer.Body.String())
	}
	var problemas, orcamentos int
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM problema_ordem_servico WHERE ordem_servico_id=$1", osDiagnosticoID).Scan(&problemas); err != nil || problemas != 2 {
		t.Fatalf("problemas=%d erro=%v", problemas, err)
	}
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM orcamento WHERE ordem_servico_id=$1 AND tipo_orcamento='PRINCIPAL' AND status='CRIADO'", osDiagnosticoID).Scan(&orcamentos); err != nil || orcamentos != 1 {
		t.Fatalf("orcamentos=%d erro=%v", orcamentos, err)
	}
	if writer := request("85000000-0000-0000-0000-000000000001", `{"descricao":"x"}`); writer.Code != http.StatusNotFound {
		t.Fatalf("os ausente: %d", writer.Code)
	}
	if writer := request(osFechadaID, `{"descricao":"x"}`); writer.Code != http.StatusConflict {
		t.Fatalf("status invalido: %d", writer.Code)
	}
}
