package ordemservico_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

const (
	clienteAna     = "20000000-0000-0000-0000-000000000001"
	veiculoAna     = "30000000-0000-0000-0000-000000000001"
	veiculoOutro   = "30000000-0000-0000-0000-000000000002"
	clienteAusente = "20000000-0000-0000-0000-999999999999"
)

type criarResponse struct {
	OrdemServicoID string    `json:"ordemServicoId"`
	ClienteID      string    `json:"clienteId"`
	VeiculoID      string    `json:"veiculoId"`
	Status         string    `json:"status"`
	CriadaEm       time.Time `json:"criadaEm"`
}

func TestCriarOrdemServico(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := presentation.NewCriarHandler(application.NewCriar(infrastructure.NewPostgresRepository(db)), jwt)

	request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(`{"clienteId":"`+clienteAna+`","veiculoId":"`+veiculoAna+`"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var created criarResponse
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(ctx, `DELETE FROM ordem_servico WHERE id = $1`, created.OrdemServicoID)
	if created.ClienteID != clienteAna || created.VeiculoID != veiculoAna || created.Status != "RECEBIDA" || created.CriadaEm.IsZero() {
		t.Fatalf("resposta: %+v", created)
	}

	var placa, status string
	var orcamentos int
	if err := db.QueryRow(ctx, `SELECT placa_veiculo, status FROM ordem_servico WHERE id = $1`, created.OrdemServicoID).Scan(&placa, &status); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM orcamento WHERE ordem_servico_id = $1`, created.OrdemServicoID).Scan(&orcamentos); err != nil {
		t.Fatal(err)
	}
	if placa != "ABC1D23" || status != "RECEBIDA" || orcamentos != 0 {
		t.Fatalf("placa = %q, status = %q, orçamentos = %d", placa, status, orcamentos)
	}

	assertStatus(t, handler, token, `{"clienteId":"`+clienteAusente+`","veiculoId":"`+veiculoAna+`"}`, http.StatusNotFound)
	assertStatus(t, handler, token, `{"clienteId":"`+clienteAna+`","veiculoId":"`+veiculoOutro+`"}`, http.StatusConflict)
}

func assertStatus(t *testing.T, handler http.Handler, token, body string, expected int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("status %d, esperado %d: %s", response.Code, expected, response.Body.String())
	}
}
