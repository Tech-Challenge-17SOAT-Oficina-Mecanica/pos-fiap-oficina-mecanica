package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	var dataFilaPendente *time.Time
	var versaoPendente int
	if err := db.QueryRow(ctx, "SELECT status, data_entrada_fila, version FROM ordem_servico WHERE id = $1", osID).Scan(&statusOS, &dataFilaPendente, &versaoPendente); err != nil || statusOS != "AGUARDANDO_RECURSOS" || dataFilaPendente != nil || versaoPendente != 2 {
		t.Fatalf("os=%s dataFila=%v version=%d err=%v", statusOS, dataFilaPendente, versaoPendente, err)
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

	const (
		osProntaID        = "95000000-0000-0000-0000-000000000009"
		orcamentoProntoID = "95000000-0000-0000-0000-000000000010"
		pecaProntaID      = "95000000-0000-0000-0000-000000000011"
	)
	defer func() {
		for _, comando := range []string{
			"DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id = $1",
			"DELETE FROM movimentacao_estoque WHERE ordem_servico_id = $1",
			"DELETE FROM reserva_estoque WHERE ordem_servico_item_id IN (SELECT id FROM ordem_servico_item WHERE ordem_servico_id = $1)",
			"DELETE FROM orcamento_item WHERE orcamento_id = $1",
			"DELETE FROM ordem_servico_item WHERE ordem_servico_id = $1",
			"DELETE FROM orcamento WHERE id = $1",
			"DELETE FROM ordem_servico WHERE id = $1",
			"DELETE FROM item_estoque WHERE id = $1",
		} {
			id := osProntaID
			if strings.Contains(comando, "orcamento_item") || strings.Contains(comando, "orcamento WHERE") {
				id = orcamentoProntoID
			}
			if strings.Contains(comando, "item_estoque") {
				id = pecaProntaID
			}
			if _, err := db.Exec(ctx, comando, id); err != nil {
				t.Errorf("cleanup pronto: %v", err)
			}
		}
	}()
	if _, err := db.Exec(ctx, `
		INSERT INTO item_estoque (id, categoria_id, tipo, codigo, nome, descricao, descricao_normalizada, fornecedor_id, unidade_medida, saldo_fisico, saldo_reservado, preco_venda, ativo)
		VALUES ($1, $2, 'PECA', 'PEC-950002', 'Peca pronta', 'Peca pronta', 'peca pronta', $3, 'UN', 1, 0, 50, TRUE);
		INSERT INTO ordem_servico (id, cliente_id, veiculo_id, placa_veiculo, status) VALUES ($4, $5, $6, 'APR1A23', 'AGUARDANDO_APROVACAO');
		INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status) VALUES ($7, $4, 'PRINCIPAL', 'CRIADO');
		INSERT INTO orcamento_item (orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
		VALUES ($7, $1, 'PECA', 'Peca pronta', 1, 50, 50);`,
		pgx.QueryExecModeSimpleProtocol, pecaProntaID, categoriaID, fornecedorID, osProntaID, clienteID, veiculoID, orcamentoProntoID); err != nil {
		t.Fatal(err)
	}
	tokenPronta, _ := jwt.GerarCliente(clienteID, osProntaID)
	respostaPronta := postAprovacao(handler, orcamentoProntoID, tokenPronta)
	if respostaPronta.Code != http.StatusOK || !strings.Contains(respostaPronta.Body.String(), `"statusOrdemServico":"AGUARDANDO_EXECUCAO"`) || !strings.Contains(respostaPronta.Body.String(), `"dataEntradaFila"`) {
		t.Fatalf("aprovacao pronta: status=%d body=%s", respostaPronta.Code, respostaPronta.Body.String())
	}
	var dataEntradaFila time.Time
	var versao int
	if err := db.QueryRow(ctx, "SELECT data_entrada_fila, version FROM ordem_servico WHERE id = $1", osProntaID).Scan(&dataEntradaFila, &versao); err != nil || dataEntradaFila.IsZero() || versao != 2 {
		t.Fatalf("fila=%v version=%d err=%v", dataEntradaFila, versao, err)
	}

	const (
		usuarioMecanicoID       = "95000000-0000-0000-0000-000000000012"
		mecanicoResponsavelID   = "95000000-0000-0000-0000-000000000013"
		osComplementarID        = "95000000-0000-0000-0000-000000000014"
		orcamentoPrincipalID    = "95000000-0000-0000-0000-000000000015"
		orcamentoComplementarID = "95000000-0000-0000-0000-000000000016"
		pecaComplementarID      = "95000000-0000-0000-0000-000000000017"
	)
	defer func() {
		comandos := []struct {
			query string
			args  []any
		}{
			{"DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id = $1", []any{osComplementarID}},
			{"DELETE FROM movimentacao_estoque WHERE ordem_servico_id = $1", []any{osComplementarID}},
			{"DELETE FROM reserva_estoque WHERE ordem_servico_item_id IN (SELECT id FROM ordem_servico_item WHERE ordem_servico_id = $1)", []any{osComplementarID}},
			{"DELETE FROM orcamento_item WHERE orcamento_id IN ($1, $2)", []any{orcamentoPrincipalID, orcamentoComplementarID}},
			{"DELETE FROM ordem_servico_item WHERE ordem_servico_id = $1", []any{osComplementarID}},
			{"DELETE FROM orcamento WHERE id IN ($1, $2)", []any{orcamentoPrincipalID, orcamentoComplementarID}},
			{"DELETE FROM ordem_servico WHERE id = $1", []any{osComplementarID}},
			{"DELETE FROM item_estoque WHERE id = $1", []any{pecaComplementarID}},
			{"DELETE FROM mecanico WHERE id = $1", []any{mecanicoResponsavelID}},
			{"DELETE FROM usuario WHERE id = $1", []any{usuarioMecanicoID}},
		}
		for _, comando := range comandos {
			if _, err := db.Exec(ctx, comando.query, comando.args...); err != nil {
				t.Errorf("cleanup complementar: %v", err)
			}
		}
	}()
	if _, err := db.Exec(ctx, `
		INSERT INTO usuario (id, email, senha_hash, ativo) VALUES ($1, 'mecanico-aprovacao@example.com', 'hash', TRUE);
		INSERT INTO mecanico (id, usuario_id, nome, version) VALUES ($2, $1, 'Mecanico aprovacao', 1);
		INSERT INTO item_estoque (id, categoria_id, tipo, codigo, nome, descricao, descricao_normalizada, fornecedor_id, unidade_medida, saldo_fisico, saldo_reservado, preco_venda, ativo)
		VALUES ($3, $4, 'PECA', 'PEC-950003', 'Peca complementar', 'Peca complementar', 'peca complementar', $5, 'UN', 1, 0, 50, TRUE);
		INSERT INTO ordem_servico (id, cliente_id, veiculo_id, mecanico_responsavel_id, placa_veiculo, status)
		VALUES ($6, $7, $8, $2, 'APR1A23', 'AGUARDANDO_APROVACAO');
		INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status) VALUES ($9, $6, 'PRINCIPAL', 'APROVADO');
		INSERT INTO orcamento (id, ordem_servico_id, orcamento_original_id, tipo_orcamento, status) VALUES ($10, $6, $9, 'COMPLEMENTAR', 'CRIADO');
		INSERT INTO orcamento_item (orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
		VALUES ($10, $3, 'PECA', 'Peca complementar', 1, 50, 50);`,
		pgx.QueryExecModeSimpleProtocol, usuarioMecanicoID, mecanicoResponsavelID, pecaComplementarID, categoriaID, fornecedorID,
		osComplementarID, clienteID, veiculoID, orcamentoPrincipalID, orcamentoComplementarID); err != nil {
		t.Fatal(err)
	}
	tokenComplementar, _ := jwt.GerarCliente(clienteID, osComplementarID)
	respostaComplementar := postAprovacao(handler, orcamentoComplementarID, tokenComplementar)
	if respostaComplementar.Code != http.StatusOK || !strings.Contains(respostaComplementar.Body.String(), `"statusOrdemServico":"AGUARDANDO_EXECUCAO"`) || !strings.Contains(respostaComplementar.Body.String(), `"dataEntradaFila"`) {
		t.Fatalf("aprovacao complementar: status=%d body=%s", respostaComplementar.Code, respostaComplementar.Body.String())
	}
	var mecanicoApos string
	if err := db.QueryRow(ctx, "SELECT data_entrada_fila, version, mecanico_responsavel_id::text FROM ordem_servico WHERE id = $1", osComplementarID).Scan(&dataEntradaFila, &versao, &mecanicoApos); err != nil || dataEntradaFila.IsZero() || versao != 2 || mecanicoApos != mecanicoResponsavelID {
		t.Fatalf("fila=%v version=%d mecanico=%q err=%v", dataEntradaFila, versao, mecanicoApos, err)
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
