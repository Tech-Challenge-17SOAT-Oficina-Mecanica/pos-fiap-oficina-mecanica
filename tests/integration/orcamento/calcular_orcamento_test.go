package orcamento_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestCalcularOrcamento(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	var configuracaoExiste bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('configuracao_oficina') IS NOT NULL`).Scan(&configuracaoExiste); err != nil || !configuracaoExiste {
		t.Skip("migration de cálculo indisponível")
	}
	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"orcamentos:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := presentation.NewCalcularHandler(application.NewCalcular(infrastructure.NewPostgresRepository(db)), jwt)
	const criado = "74000000-0000-0000-0000-000000000002"
	const aprovado = "74000000-0000-0000-0000-000000000001"
	defer db.Exec(ctx, `DELETE FROM auditoria_ordem_servico WHERE agregado_id = $1 AND tipo_evento = 'ORCAMENTO_CALCULADO'`, criado)
	cases := []struct {
		name, id, auth, contains string
		status                   int
	}{
		{"complementar", criado, "Bearer " + token, `"valorTotalGeral":754.00`, http.StatusOK},
		{"id invalido", "invalido", "Bearer " + token, "orcamentoId inválido", http.StatusBadRequest},
		{"sem token", criado, "", "token", http.StatusUnauthorized},
		{"sem escopo", criado, "Bearer " + gerarToken(t, jwt, []string{"orcamentos:ler"}), "orcamentos:escrever", http.StatusForbidden},
		{"nao encontrado", "74000000-0000-0000-0000-000000000099", "Bearer " + token, "não encontrado", http.StatusNotFound},
		{"aprovado", aprovado, "Bearer " + token, "CRIADO", http.StatusConflict},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+test.id+"/calcular", nil)
			request.SetPathValue("orcamentoId", test.id)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			handler(response, request)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.contains) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	var estimativa int
	if err := db.QueryRow(ctx, `SELECT estimativa_entrega_dias FROM orcamento WHERE id = $1`, criado).Scan(&estimativa); err != nil || estimativa != 3 {
		t.Fatalf("estimativa=%d erro=%v", estimativa, err)
	}
	var auditorias int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM auditoria_ordem_servico WHERE agregado_id = $1 AND tipo_evento = 'ORCAMENTO_CALCULADO'`, criado).Scan(&auditorias); err != nil || auditorias == 0 {
		t.Fatalf("auditorias=%d erro=%v", auditorias, err)
	}
}

func TestCalcularOrcamentoFazRollback(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	var configuracaoExiste bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('configuracao_oficina') IS NOT NULL`).Scan(&configuracaoExiste); err != nil || !configuracaoExiste {
		t.Skip("migration de cálculo indisponível")
	}
	const orcamentoID = "74000000-0000-0000-0000-000000000002"
	const itemID = "75000000-0000-0000-0000-000000000004"
	if _, err := db.Exec(ctx, `UPDATE orcamento SET estimativa_entrega_dias = 99 WHERE id = $1`, orcamentoID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE orcamento_item SET valor_total = 1 WHERE id = $1`, itemID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(ctx, `UPDATE orcamento SET estimativa_entrega_dias = 3 WHERE id = $1`, orcamentoID)
		_, _ = db.Exec(ctx, `UPDATE orcamento_item SET valor_total = 450 WHERE id = $1`, itemID)
	}()

	_, err = infrastructure.NewPostgresRepository(db).Calcular(ctx, orcamentoID, "usuario-invalido")
	if err == nil {
		t.Fatal("era esperada falha na auditoria")
	}
	var estimativa int
	var total string
	if err := db.QueryRow(ctx, `SELECT o.estimativa_entrega_dias, oi.valor_total::text FROM orcamento o JOIN orcamento_item oi ON oi.orcamento_id = o.id WHERE o.id = $1 AND oi.id = $2`, orcamentoID, itemID).Scan(&estimativa, &total); err != nil {
		t.Fatal(err)
	}
	if estimativa != 99 || total != "1.00" {
		t.Fatalf("rollback não preservou dados: estimativa=%d total=%s", estimativa, total)
	}
}

func gerarToken(t *testing.T, jwt segurancaInfrastructure.JWT, escopos []string) string {
	t.Helper()
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", escopos)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
