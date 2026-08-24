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
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestDesativarEReativarFornecedor(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"compras:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	semPermissao, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"compras:ler"})
	if err != nil {
		t.Fatal(err)
	}

	repository := infrastructure.NewPostgresRepository(db)
	mux := http.NewServeMux()
	mux.Handle("DELETE /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", presentation.NewDesativarHandler(application.NewDesativarFornecedor(repository))))
	mux.Handle("POST /fornecedores/{fornecedorId}/reativacao", segurancaPresentation.RequireScope(jwt, "compras:escrever", presentation.NewReativarHandler(application.NewReativarFornecedor(repository))))

	fornecedorID := inserirFornecedorSituacao(t, ctx, db, "Fornecedor Situacao Ltda")

	request := httptest.NewRequest(http.MethodDelete, "/fornecedores/"+fornecedorID, nil)
	request.SetPathValue("fornecedorId", fornecedorID)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}

	var ativo bool
	if err := db.QueryRow(ctx, "SELECT ativo FROM fornecedor WHERE id = $1", fornecedorID).Scan(&ativo); err != nil || ativo {
		t.Fatalf("ativo=%v err=%v", ativo, err)
	}

	request = httptest.NewRequest(http.MethodDelete, "/fornecedores/"+fornecedorID, nil)
	request.SetPathValue("fornecedorId", fornecedorID)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status inativo=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/fornecedores/"+fornecedorID+"/reativacao", nil)
	request.SetPathValue("fornecedorId", fornecedorID)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status reativacao=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/fornecedores/"+fornecedorID+"/reativacao", nil)
	request.SetPathValue("fornecedorId", fornecedorID)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("status ativo=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/fornecedores/"+fornecedorID, nil)
	request.SetPathValue("fornecedorId", fornecedorID)
	request.Header.Set("Authorization", "Bearer "+semPermissao)
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status sem escopo=%d body=%s", response.Code, response.Body.String())
	}
}

func TestDesativarFornecedorComPedidoAberto(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"compras:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewPostgresRepository(db)
	mux := http.NewServeMux()
	mux.Handle("DELETE /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", presentation.NewDesativarHandler(application.NewDesativarFornecedor(repository))))

	fornecedorID := inserirFornecedorSituacao(t, ctx, db, "Fornecedor Bloqueado Ltda")
	pedidoID := inserirPedidoCompra(t, ctx, db, fornecedorID, "ABERTO")
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra WHERE id = $1", pedidoID)
	})

	request := httptest.NewRequest(http.MethodDelete, "/fornecedores/"+fornecedorID, nil)
	request.SetPathValue("fornecedorId", fornecedorID)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func inserirFornecedorSituacao(t *testing.T, ctx context.Context, db *pgxpool.Pool, razaoSocial string) string {
	t.Helper()
	documento := cnpjValido(fmt.Sprintf("%012d", time.Now().UnixNano()%1000000000000))
	var fornecedorID string
	err := db.QueryRow(ctx, `
		INSERT INTO fornecedor (razao_social, documento, tipo_documento, telefone, prazo_entrega_dias)
		VALUES ($1, $2, 'CNPJ', '11999990000', 7)
		RETURNING id`, razaoSocial, documento).Scan(&fornecedorID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM fornecedor WHERE id = $1", fornecedorID)
	})
	return fornecedorID
}

func inserirPedidoCompra(t *testing.T, ctx context.Context, db *pgxpool.Pool, fornecedorID, status string) string {
	t.Helper()
	var pedidoID string
	numero := fmt.Sprintf("TESTE-%d", time.Now().UnixNano())
	err := db.QueryRow(ctx, `
		INSERT INTO pedido_compra (fornecedor_id, numero, status)
		VALUES ($1, $2, $3)
		RETURNING id`, fornecedorID, numero, status).Scan(&pedidoID)
	if err != nil {
		t.Fatal(err)
	}
	return pedidoID
}
