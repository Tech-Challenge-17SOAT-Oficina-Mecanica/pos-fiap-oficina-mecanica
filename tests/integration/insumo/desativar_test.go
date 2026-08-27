package insumo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/insumo"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/insumo"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestDesativarInsumo(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const (
		id              = "5f000000-0000-0000-0000-000000000001"
		orcamentoItemID = "5f000000-0000-0000-0000-000000000002"
		usuarioID       = "90000000-0000-0000-0000-000000000001"
	)
	_, _ = db.Exec(ctx, `DELETE FROM orcamento_item WHERE id = $1`, orcamentoItemID)
	_, _ = db.Exec(ctx, `DELETE FROM item_estoque WHERE id = $1`, id)
	defer func() {
		_, _ = db.Exec(ctx, `DELETE FROM orcamento_item WHERE id = $1`, orcamentoItemID)
		_, _ = db.Exec(ctx, `DELETE FROM item_estoque WHERE id = $1`, id)
	}()
	_, err = db.Exec(ctx, `INSERT INTO item_estoque
		(id, categoria_id, tipo, codigo, nome, descricao, descricao_normalizada, unidade_medida,
		saldo_fisico, saldo_reservado, estoque_minimo, custo_unitario, ativo, version)
		VALUES ($1, '10000000-0000-0000-0000-000000000002', 'INSUMO', 'INS-999991',
		'Oleo Integracao', 'Oleo integracao', 'oleo integracao', 'L', 5.500, 0, 1, 20, TRUE, 1)`, id)
	if err != nil {
		t.Fatal(err)
	}

	jwt, _ := segurancaInfrastructure.NewJWT("segredo-de-teste")
	token, _ := jwt.Gerar(usuarioID, []string{"estoque:escrever"})
	semEscopo, _ := jwt.Gerar(usuarioID, []string{"estoque:ler"})
	repository := infrastructure.NewPostgresRepository(db)
	handler := segurancaPresentation.RequireScope(jwt, "estoque:escrever",
		presentation.NewDesativarHandler(application.NewDesativarInsumo(repository)))

	_, err = db.Exec(ctx, `INSERT INTO orcamento_item
		(id, orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
		VALUES ($1, '74000000-0000-0000-0000-000000000002', $2, 'INSUMO', 'Oleo integracao', 1, 20, 20)`, orcamentoItemID, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := executarDeleteInsumo(handler, id, token); got.Code != http.StatusConflict ||
		!strings.Contains(got.Body.String(), "70000000-0000-0000-0000-000000000001") {
		t.Fatalf("orcamento criado status=%d body=%s", got.Code, got.Body.String())
	}
	if _, err := db.Exec(ctx, `DELETE FROM orcamento_item WHERE id = $1`, orcamentoItemID); err != nil {
		t.Fatal(err)
	}

	response := executarDeleteInsumo(handler, id, token)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ativo":false`) ||
		!strings.Contains(response.Body.String(), `"usuarioDesativacao":"`+usuarioID+`"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var quantidade int
	var ativo bool
	var saldo string
	var usuario string
	if err := db.QueryRow(ctx, `SELECT COUNT(*), BOOL_AND(ativo), MAX(saldo_fisico::text), MAX(usuario_desativacao::text)
		FROM item_estoque WHERE id = $1`, id).Scan(&quantidade, &ativo, &saldo, &usuario); err != nil ||
		quantidade != 1 || ativo || saldo != "5.500" || usuario != usuarioID {
		t.Fatalf("quantidade=%d ativo=%v saldo=%q usuario=%q err=%v", quantidade, ativo, saldo, usuario, err)
	}
	if got := executarDeleteInsumo(handler, id, token); got.Code != http.StatusConflict {
		t.Fatalf("segunda exclusao status=%d body=%s", got.Code, got.Body.String())
	}
	if got := executarDeleteInsumo(handler, "50000000-0000-0000-0000-000000000003", token); got.Code != http.StatusConflict ||
		!strings.Contains(got.Body.String(), "70000000-0000-0000-0000-000000000001") {
		t.Fatalf("reservado status=%d body=%s", got.Code, got.Body.String())
	}
	if got := executarDeleteInsumo(handler, "5f000000-0000-0000-0000-0000000000ff", token); got.Code != http.StatusNotFound {
		t.Fatalf("nao encontrado status=%d", got.Code)
	}
	if got := executarDeleteInsumo(handler, id, semEscopo); got.Code != http.StatusForbidden {
		t.Fatalf("sem escopo status=%d", got.Code)
	}
	if got := executarDeleteInsumo(handler, "abc", token); got.Code != http.StatusBadRequest {
		t.Fatalf("uuid status=%d", got.Code)
	}
}

func executarDeleteInsumo(handler http.Handler, id, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodDelete, "/estoque/insumos/"+id, nil)
	request.SetPathValue("insumoId", id)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
