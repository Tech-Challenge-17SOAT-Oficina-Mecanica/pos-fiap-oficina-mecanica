package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestRegistrarEntregaDoVeiculo(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	osID := id("b1000000-0000-0000-0000-")
	osSemValorID := id("b2000000-0000-0000-0000-")
	orcamentoID := id("b3000000-0000-0000-0000-")
	itemID := id("b4000000-0000-0000-0000-")

	_, err = db.Exec(ctx, `
		INSERT INTO ordem_servico (id, cliente_id, veiculo_id, placa_veiculo, status, finalizada_em)
		VALUES ($1, $3, $4, 'ABC1D23', 'FINALIZADA', CURRENT_TIMESTAMP),
		       ($2, $3, $4, 'ABC1D23', 'FINALIZADA', CURRENT_TIMESTAMP)`,
		osID, osSemValorID, clienteAna, veiculoAna)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status, aprovado_em)
		VALUES ($1, $2, 'PRINCIPAL', 'APROVADO', CURRENT_TIMESTAMP)`, orcamentoID, osID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `INSERT INTO orcamento_item
		(id, orcamento_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
		VALUES ($1, $2, 'SERVICO', 'Servico de teste', 1, 199.90, 199.90)`, itemID, orcamentoID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, `DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id IN ($1, $2)`, osID, osSemValorID)
		_, _ = db.Exec(ctx, `DELETE FROM orcamento_item WHERE id = $1`, itemID)
		_, _ = db.Exec(ctx, `DELETE FROM orcamento WHERE id = $1`, orcamentoID)
		_, _ = db.Exec(ctx, `DELETE FROM ordem_servico WHERE id IN ($1, $2)`, osID, osSemValorID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "os:escrever",
		presentation.NewEntregarHandler(application.NewEntregar(infrastructure.NewPostgresRepository(db))))

	request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/entrega",
		strings.NewReader(`{"clienteId":"`+clienteAna+`","observacoes":" sem ressalvas "}`))
	request.SetPathValue("osId", osID)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var result struct {
		Status     string  `json:"status"`
		ValorFinal float64 `json:"valorFinal"`
	}
	if err = json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != domain.StatusEntregue || result.ValorFinal != 199.90 {
		t.Fatalf("resposta=%+v", result)
	}

	var status, observacoes string
	var valorFinal float64
	var auditorias int
	if err = db.QueryRow(ctx, `SELECT status, valor_final, observacoes_entrega FROM ordem_servico WHERE id=$1`, osID).
		Scan(&status, &valorFinal, &observacoes); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(ctx, `SELECT COUNT(*) FROM auditoria_ordem_servico WHERE ordem_servico_id=$1 AND tipo_evento='VEICULO_ENTREGUE'`, osID).
		Scan(&auditorias); err != nil {
		t.Fatal(err)
	}
	if status != domain.StatusEntregue || valorFinal != 199.90 || observacoes != "sem ressalvas" || auditorias != 1 {
		t.Fatalf("status=%s valor=%.2f observacoes=%q auditorias=%d", status, valorFinal, observacoes, auditorias)
	}

	assertEntregaStatus(t, handler, token, osID, `{}`, http.StatusConflict)
	assertEntregaStatus(t, handler, token, osSemValorID, `{}`, http.StatusConflict)
	assertEntregaStatus(t, handler, "", osSemValorID, `{}`, http.StatusUnauthorized)

	tokenSemEscopo, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"os:ler"})
	if err != nil {
		t.Fatal(err)
	}
	assertEntregaStatus(t, handler, tokenSemEscopo, osSemValorID, `{}`, http.StatusForbidden)
}

func assertEntregaStatus(t *testing.T, handler http.Handler, token, osID, body string, expected int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/entrega", strings.NewReader(body))
	request.SetPathValue("osId", osID)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("status=%d esperado=%d body=%s", response.Code, expected, response.Body.String())
	}
}
