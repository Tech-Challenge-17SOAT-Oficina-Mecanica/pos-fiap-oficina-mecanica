package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	segurancaDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	orcamentoInfra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	segurancaInfra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	orcamentoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestAprovarOrcamento(t *testing.T) {
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	if _, err := db.Exec(ctx, `
		CREATE SEQUENCE IF NOT EXISTS seq_pedido_compra_numero START WITH 1;
		ALTER TABLE orcamento ADD COLUMN IF NOT EXISTS cliente_aprovador_id UUID REFERENCES cliente (id);`, pgx.QueryExecModeSimpleProtocol); err != nil {
		t.Fatal(err)
	}

	const (
		clienteID    = "95000000-0000-0000-0000-000000000001"
		veiculoID    = "95000000-0000-0000-0000-000000000002"
		osID         = "95000000-0000-0000-0000-000000000003"
		fornecedorID = "95000000-0000-0000-0000-000000000004"
		categoriaID  = "95000000-0000-0000-0000-000000000005"
		pecaID       = "95000000-0000-0000-0000-000000000006"
		orcamentoID  = "95000000-0000-0000-0000-000000000007"
		outroOSID    = "95000000-0000-0000-0000-000000000008"
	)
	cleanupAprovacao(ctx, t, db, osID, outroOSID, orcamentoID, clienteID, veiculoID, fornecedorID, categoriaID, pecaID)
	defer cleanupAprovacao(ctx, t, db, osID, outroOSID, orcamentoID, clienteID, veiculoID, fornecedorID, categoriaID, pecaID)

	if _, err := db.Exec(ctx, `
		INSERT INTO categoria (id, nome) VALUES ($1, 'Aprovacao Orcamento') ON CONFLICT (id) DO NOTHING;
		INSERT INTO cliente (id, nome, documento, tipo_documento, telefone, ativo) VALUES ($2, 'Cliente Aprovacao', '12345678904', 'CPF', '11999999996', TRUE);
		INSERT INTO veiculo (id, cliente_id, placa, marca, modelo, ano, ativo) VALUES ($3, $2, 'APR1A23', 'VW', 'Gol', 2020, TRUE);
		INSERT INTO fornecedor (id, razao_social, documento, tipo_documento, ativo) VALUES ($4, 'Fornecedor Aprovacao', '12345678000196', 'CNPJ', TRUE);
		INSERT INTO item_estoque (id, categoria_id, tipo, codigo, nome, descricao, descricao_normalizada, fornecedor_id, unidade_medida, saldo_fisico, saldo_reservado, preco_venda, ativo)
		VALUES ($5, $1, 'PECA', 'PEC-950001', 'Peca aprovacao', 'Peca aprovacao', 'peca aprovacao', $4, 'UN', 1, 0, 50, TRUE);
		INSERT INTO ordem_servico (id, cliente_id, veiculo_id, placa_veiculo, status) VALUES ($6, $2, $3, 'APR1A23', 'AGUARDANDO_APROVACAO');
		INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status) VALUES ($7, $6, 'PRINCIPAL', 'CRIADO');
		INSERT INTO orcamento_item (orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
		VALUES ($7, $5, 'PECA', 'Peca aprovacao', 2, 50, 100);`,
		pgx.QueryExecModeSimpleProtocol, categoriaID, clienteID, veiculoID, fornecedorID, pecaID, osID, orcamentoID); err != nil {
		t.Fatal(err)
	}

	jwt, err := segurancaInfra.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.GerarCliente(clienteID, osID)
	tokenOutraOS, _ := jwt.GerarCliente(clienteID, outroOSID)
	handler := segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOrcamentosDecidir,
		orcamentoPresentation.NewAprovarHandler(application.NewAprovar(orcamentoInfra.NewPostgresRepository(db))))

	if response := postAprovacao(handler, orcamentoID, ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", response.Code)
	}
	if response := postAprovacao(handler, orcamentoID, tokenOutraOS); response.Code != http.StatusForbidden {
		t.Fatalf("outra os=%d body=%s", response.Code, response.Body.String())
	}
	response := postAprovacao(handler, orcamentoID, token)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"statusOrdemServico":"AGUARDANDO_RECURSOS"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if replay := postAprovacao(handler, orcamentoID, token); replay.Code != http.StatusConflict {
		t.Fatalf("replay=%d body=%s", replay.Code, replay.Body.String())
	}

	var statusOrcamento, statusOS, aprovador string
	if err := db.QueryRow(ctx, "SELECT status, cliente_aprovador_id::text FROM orcamento WHERE id = $1", orcamentoID).Scan(&statusOrcamento, &aprovador); err != nil {
		t.Fatal(err)
	}
	if statusOrcamento != "APROVADO" || aprovador != clienteID {
		t.Fatalf("orcamento=%s aprovador=%s", statusOrcamento, aprovador)
	}
	if err := db.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1", osID).Scan(&statusOS); err != nil || statusOS != "AGUARDANDO_RECURSOS" {
		t.Fatalf("os=%s err=%v", statusOS, err)
	}
	var saldoReservado string
	if err := db.QueryRow(ctx, "SELECT saldo_reservado::text FROM item_estoque WHERE id = $1", pecaID).Scan(&saldoReservado); err != nil || saldoReservado != "1.000" {
		t.Fatalf("saldo=%s err=%v", saldoReservado, err)
	}
	var reservas int
	var quantidadePedida string
	if err := db.QueryRow(ctx, `
		SELECT pci.quantidade_pedida::text
		FROM pedido_compra pc
		JOIN pedido_compra_item pci ON pci.pedido_compra_id = pc.id
		JOIN pedido_compra_item_os pcios ON pcios.pedido_compra_item_id = pci.id
		JOIN ordem_servico_item osi ON osi.id = pcios.ordem_servico_item_id
		WHERE pci.item_estoque_id = $1 AND osi.ordem_servico_id = $2 AND pc.fornecedor_id = $3`,
		pecaID, osID, fornecedorID).Scan(&quantidadePedida); err != nil || quantidadePedida != "1.000" {
		t.Fatalf("quantidadePedida=%s err=%v", quantidadePedida, err)
	}
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM reserva_estoque r JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id WHERE osi.ordem_servico_id = $1 AND r.status = 'ATIVA'", osID).Scan(&reservas); err != nil || reservas != 1 {
		t.Fatalf("reservas=%d err=%v", reservas, err)
	}
}

func postAprovacao(handler http.Handler, orcamentoID, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+orcamentoID+"/aprovar", nil)
	request.SetPathValue("orcamentoId", orcamentoID)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func cleanupAprovacao(ctx context.Context, t *testing.T, db *pgxpool.Pool, osID, outroOSID, orcamentoID, clienteID, veiculoID, fornecedorID, categoriaID, pecaID string) {
	t.Helper()
	comandos := []string{
		"DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id IN ($1,$2)",
		"DELETE FROM movimentacao_estoque WHERE ordem_servico_id = $1",
		"DELETE FROM reserva_estoque WHERE ordem_servico_item_id IN (SELECT id FROM ordem_servico_item WHERE ordem_servico_id = $1)",
		"DELETE FROM pedido_compra_item_os WHERE ordem_servico_item_id IN (SELECT id FROM ordem_servico_item WHERE ordem_servico_id = $1)",
		"DELETE FROM pedido_compra_item WHERE item_estoque_id = $1",
		"DELETE FROM pedido_compra WHERE fornecedor_id = $1",
		"DELETE FROM orcamento_item WHERE orcamento_id = $1",
		"DELETE FROM ordem_servico_item WHERE ordem_servico_id = $1",
		"DELETE FROM orcamento WHERE id = $1",
		"DELETE FROM ordem_servico WHERE id IN ($1,$2)",
		"DELETE FROM item_estoque WHERE id = $1",
		"DELETE FROM fornecedor WHERE id = $1",
		"DELETE FROM veiculo WHERE id = $1",
		"DELETE FROM cliente WHERE id = $1",
		"DELETE FROM categoria WHERE id = $1",
	}
	args := [][]any{{osID, outroOSID}, {osID}, {osID}, {osID}, {pecaID}, {fornecedorID}, {orcamentoID}, {osID}, {orcamentoID}, {osID, outroOSID}, {pecaID}, {fornecedorID}, {veiculoID}, {clienteID}, {categoriaID}}
	for indice, comando := range comandos {
		if _, err := db.Exec(ctx, comando, args[indice]...); err != nil {
			t.Fatalf("cleanup %q: %v", comando, err)
		}
	}
}
