package integration_test

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
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

const osTeste = "76000000-0000-0000-0000-000000000001"

func TestRegistrarProblemaRelatado(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	_, _ = db.Exec(ctx, `DELETE FROM problema_ordem_servico WHERE ordem_servico_id = $1`, osTeste)
	_, _ = db.Exec(ctx, `DELETE FROM ordem_servico WHERE id = $1`, osTeste)
	_, err = db.Exec(ctx, `INSERT INTO ordem_servico (id, cliente_id, veiculo_id, placa_veiculo, status)
		VALUES ($1, '20000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'ABC1D23', 'RECEBIDA')`, osTeste)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM problema_ordem_servico WHERE ordem_servico_id = $1`, osTeste)
		_, _ = db.Exec(ctx, `DELETE FROM ordem_servico WHERE id = $1`, osTeste)
	}()

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "os:escrever", presentation.NewRegistrarProblemaRelatadoHandler(application.NewRegistrarProblemaRelatado(infrastructure.NewPostgresRepository(db))))
	semEscopo, err := jwt.Gerar("usuario", []string{"os:ler"})
	if err != nil {
		t.Fatal(err)
	}
	assertProblemaRelatadoStatus(t, handler, osTeste, "", `{"descricao":"Ruído"}`, http.StatusUnauthorized)
	assertProblemaRelatadoStatus(t, handler, osTeste, semEscopo, `{"descricao":"Ruído"}`, http.StatusForbidden)
	assertProblemaRelatadoStatus(t, handler, "76000000-0000-0000-0000-999999999999", token, `{"descricao":"Ruído"}`, http.StatusNotFound)
	assertProblemaRelatadoStatus(t, handler, "70000000-0000-0000-0000-000000000001", token, `{"descricao":"Ruído"}`, http.StatusConflict)

	request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osTeste+"/problema-relatado", strings.NewReader(`{"descricao":"Veículo apresenta ruído ao frear","observacoes":"Há uma semana"}`))
	request.SetPathValue("osId", osTeste)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var body struct {
		Status                string    `json:"status"`
		DataInicioDiagnostico time.Time `json:"dataInicioDiagnostico"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Status != "EM_DIAGNOSTICO" || body.DataInicioDiagnostico.IsZero() {
		t.Fatalf("body = %+v, erro = %v", body, err)
	}

	var status, descricao, observacoes string
	var inicio time.Time
	err = db.QueryRow(ctx, `SELECT status, data_inicio_diagnostico, problema_relatado_descricao,
		COALESCE(problema_relatado_observacoes, '') FROM ordem_servico WHERE id = $1`, osTeste).
		Scan(&status, &inicio, &descricao, &observacoes)
	if err != nil || status != "EM_DIAGNOSTICO" || inicio.IsZero() || descricao != "Veículo apresenta ruído ao frear" || observacoes != "Há uma semana" {
		t.Fatalf("status=%q início=%v descrição=%q observações=%q erro=%v", status, inicio, descricao, observacoes, err)
	}

	duplicate := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osTeste+"/problema-relatado", strings.NewReader(`{"descricao":"Outro relato"}`))
	duplicate.SetPathValue("osId", osTeste)
	duplicate.Header.Set("Authorization", "Bearer "+token)
	duplicateResponse := httptest.NewRecorder()
	handler.ServeHTTP(duplicateResponse, duplicate)
	if duplicateResponse.Code != http.StatusConflict {
		t.Fatalf("duplicado status %d: %s", duplicateResponse.Code, duplicateResponse.Body.String())
	}
}

func assertProblemaRelatadoStatus(t *testing.T, handler http.Handler, osID, token, body string, expected int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/problema-relatado", strings.NewReader(body))
	request.SetPathValue("osId", osID)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("status %d, esperado %d: %s", response.Code, expected, response.Body.String())
	}
}
