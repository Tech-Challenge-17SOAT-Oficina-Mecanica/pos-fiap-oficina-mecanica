package peca_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	segurancaDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	infra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestSolicitarCompraEReservarPecas(t *testing.T) {
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	if _, err := db.Exec(ctx, `CREATE SEQUENCE IF NOT EXISTS seq_pedido_compra_numero START WITH 1`); err != nil {
		t.Fatal(err)
	}

	const (
		clienteID    = "91000000-0000-0000-0000-000000000001"
		veiculoID    = "91000000-0000-0000-0000-000000000002"
		osID         = "91000000-0000-0000-0000-000000000003"
		fornecedorID = "91000000-0000-0000-0000-000000000004"
		categoriaID  = "91000000-0000-0000-0000-000000000005"
		pecaComSaldo = "91000000-0000-0000-0000-000000000006"
		pecaSemSaldo = "91000000-0000-0000-0000-000000000007"
		orcamentoID  = "91000000-0000-0000-0000-000000000008"
		chave        = "91000000-0000-0000-0000-000000000009"
		outraChave   = "91000000-0000-0000-0000-000000000010"
	)
	cleanupProcessamento(ctx, t, db, osID, orcamentoID, clienteID, veiculoID, fornecedorID, categoriaID, pecaComSaldo, pecaSemSaldo, chave, outraChave)
	defer cleanupProcessamento(ctx, t, db, osID, orcamentoID, clienteID, veiculoID, fornecedorID, categoriaID, pecaComSaldo, pecaSemSaldo, chave, outraChave)

	if _, err := db.Exec(ctx, `
		INSERT INTO categoria (id, nome) VALUES ($1, 'Processamento Pecas') ON CONFLICT (id) DO NOTHING;
		INSERT INTO cliente (id, nome, documento, tipo_documento, telefone, ativo) VALUES ($2, 'Cliente Processamento', '12345678901', 'CPF', '11999999999', TRUE);
		INSERT INTO veiculo (id, cliente_id, placa, marca, modelo, ano, ativo) VALUES ($3, $2, 'PRC1A23', 'VW', 'Gol', 2020, TRUE);
		INSERT INTO fornecedor (id, razao_social, documento, tipo_documento, ativo) VALUES ($4, 'Fornecedor Processamento', '12345678000199', 'CNPJ', TRUE);
		INSERT INTO item_estoque (id, categoria_id, tipo, codigo, nome, descricao, descricao_normalizada, unidade_medida, saldo_fisico, saldo_reservado, preco_venda, ativo)
		VALUES
			($5, $1, 'PECA', 'PEC-910001', 'Peca com saldo', 'Peca com saldo', 'peca com saldo', 'UN', 5, 1, 50, TRUE),
			($6, $1, 'PECA', 'PEC-910002', 'Peca sem saldo', 'Peca sem saldo', 'peca sem saldo', 'UN', 0, 0, 120, TRUE);
		INSERT INTO ordem_servico (id, cliente_id, veiculo_id, placa_veiculo, status) VALUES ($7, $2, $3, 'PRC1A23', 'AGUARDANDO_APROVACAO');
		INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status) VALUES ($8, $7, 'PRINCIPAL', 'APROVADO');
		INSERT INTO orcamento_item (orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
		VALUES
			($8, $5, 'PECA', 'Peca com saldo', 3, 50, 150),
			($8, $6, 'PECA', 'Peca sem saldo', 2, 120, 240);`,
		categoriaID, clienteID, veiculoID, fornecedorID, pecaComSaldo, pecaSemSaldo, osID, orcamentoID); err != nil {
		t.Fatal(err)
	}

	jwt, err := seguranca.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{segurancaDominio.EscopoEstoqueMovimentar})
	semEscopo, _ := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{segurancaDominio.EscopoEstoqueLer})
	handler := segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueMovimentar,
		presentation.NewSolicitarCompraEReservarPecasHandler(application.NewSolicitarCompraEReservarPecas(infra.NewPostgresRepository(db))))

	body := `{"ordemServicoId":"` + osID + `","fornecedorId":"` + fornecedorID + `","itens":[{"itemId":"` + pecaComSaldo + `","quantidade":3},{"itemId":"` + pecaSemSaldo + `","quantidade":2}]}`

	if response := postProcessamento(handler, body, "", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", response.Code)
	}
	if response := postProcessamento(handler, body, semEscopo, chave); response.Code != http.StatusForbidden {
		t.Fatalf("sem escopo=%d", response.Code)
	}

	response := postProcessamento(handler, body, token, chave)
	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"statusOrdemServico":"AGUARDANDO_RECURSOS"`) || !strings.Contains(response.Body.String(), `"pecasCompraSolicitada"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if replay := postProcessamento(handler, body, token, chave); replay.Code != http.StatusOK {
		t.Fatalf("replay=%d body=%s", replay.Code, replay.Body.String())
	}
	if duplicado := postProcessamento(handler, body, token, outraChave); duplicado.Code != http.StatusConflict {
		t.Fatalf("duplicado=%d body=%s", duplicado.Code, duplicado.Body.String())
	}

	var saldoReservado int64
	if err := db.QueryRow(ctx, "SELECT saldo_reservado FROM item_estoque WHERE id = $1", pecaComSaldo).Scan(&saldoReservado); err != nil || saldoReservado != 4 {
		t.Fatalf("saldoReservado=%d err=%v", saldoReservado, err)
	}
	var status string
	if err := db.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1", osID).Scan(&status); err != nil || status != "AGUARDANDO_RECURSOS" {
		t.Fatalf("status=%q err=%v", status, err)
	}
	var pedidos, reservas, auditorias int
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM pedido_compra pc JOIN pedido_compra_item pci ON pci.pedido_compra_id = pc.id WHERE pc.fornecedor_id = $1 AND pci.item_estoque_id = $2", fornecedorID, pecaSemSaldo).Scan(&pedidos); err != nil || pedidos != 1 {
		t.Fatalf("pedidos=%d err=%v", pedidos, err)
	}
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM reserva_estoque r JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id WHERE osi.ordem_servico_id = $1 AND r.status = 'ATIVA'", osID).Scan(&reservas); err != nil || reservas != 1 {
		t.Fatalf("reservas=%d err=%v", reservas, err)
	}
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM auditoria_ordem_servico WHERE ordem_servico_id = $1 AND tipo_evento = 'PECAS_RESERVA_COMPRA_PROCESSADA'", osID).Scan(&auditorias); err != nil || auditorias != 1 {
		t.Fatalf("auditorias=%d err=%v", auditorias, err)
	}
}

func postProcessamento(handler http.Handler, body, token, chave string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/estoque/solicitacoes-compra-reserva", strings.NewReader(body))
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if chave != "" {
		request.Header.Set("Idempotency-Key", chave)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func cleanupProcessamento(ctx context.Context, t *testing.T, db *pgxpool.Pool, osID, orcamentoID, clienteID, veiculoID, fornecedorID, categoriaID, pecaComSaldo, pecaSemSaldo, chave, outraChave string) {
	t.Helper()
	comandos := []string{
		"DELETE FROM chave_idempotencia WHERE chave IN ($1,$2)",
		"DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id = $3",
		"DELETE FROM movimentacao_estoque WHERE ordem_servico_id = $3",
		"DELETE FROM reserva_estoque WHERE ordem_servico_item_id IN (SELECT id FROM ordem_servico_item WHERE ordem_servico_id = $3)",
		"DELETE FROM pedido_compra_item_os WHERE ordem_servico_item_id IN (SELECT id FROM ordem_servico_item WHERE ordem_servico_id = $3)",
		"DELETE FROM pedido_compra_item WHERE item_estoque_id IN ($8,$9)",
		"DELETE FROM pedido_compra WHERE fornecedor_id = $6",
		"DELETE FROM orcamento_item WHERE orcamento_id = $4",
		"DELETE FROM ordem_servico_item WHERE ordem_servico_id = $3",
		"DELETE FROM orcamento WHERE id = $4",
		"DELETE FROM ordem_servico WHERE id = $3",
		"DELETE FROM item_estoque WHERE id IN ($8,$9)",
		"DELETE FROM fornecedor WHERE id = $6",
		"DELETE FROM veiculo WHERE id = $5",
		"DELETE FROM cliente WHERE id = $10",
		"DELETE FROM categoria WHERE id = $7",
	}
	args := []any{chave, outraChave, osID, orcamentoID, veiculoID, fornecedorID, categoriaID, pecaComSaldo, pecaSemSaldo, clienteID}
	for _, comando := range comandos {
		if _, err := db.Exec(ctx, comando, args...); err != nil {
			t.Fatalf("cleanup %q: %v", comando, err)
		}
	}
}
