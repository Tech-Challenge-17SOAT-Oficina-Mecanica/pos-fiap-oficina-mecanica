package integration_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/estoque"
)

func TestRegistrarEntradaEstoque(t *testing.T) {
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
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	categoriaID := id("b1000000-0000-0000-0000-")
	insumoID, pecaInativaID, pecaPedidoID := id("b2000000-0000-0000-0000-"), id("b3000000-0000-0000-0000-"), id("b4000000-0000-0000-0000-")
	fornecedorID, pedidoID, pedidoItemID := id("b5000000-0000-0000-0000-"), id("b6000000-0000-0000-0000-"), id("b7000000-0000-0000-0000-")
	clienteID, veiculoID, osID, osItemID := id("b8000000-0000-0000-0000-"), id("b9000000-0000-0000-0000-"), id("ba000000-0000-0000-0000-"), id("bb000000-0000-0000-0000-")
	codigo := func(prefix string) string {
		var bytes [3]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%s-%x", prefix, bytes)
	}
	placa := fmt.Sprintf("E%06X", suffix&0xffffff)

	if _, err = db.Exec(ctx, "INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)", categoriaID, fmt.Sprintf("Categoria %x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO item_estoque (id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado,ativo,custo_unitario) VALUES
		($1,$4,'INSUMO',$5,'Insumo entrada','Insumo entrada','insumo entrada','L',10,0,true,20.00),
		($2,$4,'PECA',$6,'Peca inativa','Peca inativa','peca inativa','UN',5,0,false,10.00),
		($3,$4,'PECA',$7,'Peca pedido','Peca pedido','peca pedido','UN',0,0,true,10.00)`,
		insumoID, pecaInativaID, pecaPedidoID, categoriaID,
		codigo("INS"), codigo("PEC"), codigo("PEC")); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO fornecedor (id,razao_social,documento,tipo_documento) VALUES ($1,'Fornecedor Teste',$2,'CNPJ')",
		fornecedorID, fmt.Sprintf("%014d", suffix%100000000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO pedido_compra (id,fornecedor_id,numero,status) VALUES ($1,$2,$3,'ABERTO')",
		pedidoID, fornecedorID, fmt.Sprintf("ENT/%012x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO pedido_compra_item (id,pedido_compra_id,item_estoque_id,quantidade_necessaria,quantidade_pedida) VALUES ($1,$2,$3,5,5)",
		pedidoItemID, pedidoID, pecaPedidoID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')",
		clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES ($1,$2,$3,$4,'AGUARDANDO_RECURSOS')", osID, clienteID, veiculoID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico_item (id,ordem_servico_id,item_estoque_id,quantidade_necessaria,valor_unitario) VALUES ($1,$2,$3,5,10.00)", osItemID, osID, pecaPedidoID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO pedido_compra_item_os (pedido_compra_item_id,ordem_servico_item_id,quantidade_atendida) VALUES ($1,$2,5)", pedidoItemID, osItemID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_estoque WHERE item_estoque_id IN ($1,$2,$3)", insumoID, pecaInativaID, pecaPedidoID)
		_, _ = db.Exec(ctx, "DELETE FROM movimentacao_estoque WHERE item_estoque_id IN ($1,$2,$3)", insumoID, pecaInativaID, pecaPedidoID)
		_, _ = db.Exec(ctx, "DELETE FROM chave_idempotencia WHERE chave IN ($1,$2)", chaveSimples(suffix), chavePedido(suffix))
		_, _ = db.Exec(ctx, "DELETE FROM reserva_estoque WHERE ordem_servico_item_id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra_item_os WHERE pedido_compra_item_id=$1", pedidoItemID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra_item WHERE id=$1", pedidoItemID)
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra WHERE id=$1", pedidoID)
		_, _ = db.Exec(ctx, "DELETE FROM fornecedor WHERE id=$1", fornecedorID)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id IN ($1,$2,$3)", insumoID, pecaInativaID, pecaPedidoID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id=$1", categoriaID)
	})

	useCase := application.NewRegistrarEntrada(infrastructure.NewPostgresRepository(db))

	// Entrada simples, sem pedido: saldo fisico sobe, custo atualiza, sem alterar reservado.
	resultado, err := useCase.Execute(ctx, application.RegistrarEntradaInput{
		IdempotencyKey: chaveSimples(suffix), DocumentoOrigem: "NF-" + fmt.Sprintf("%012x", suffix), FornecedorID: fornecedorID,
		Itens: []application.ItemInput{{ItemID: insumoID, Quantidade: 5, CustoUnitario: 25.00}},
	})
	if err != nil {
		t.Fatalf("entrada simples: %v", err)
	}
	if resultado.JaProcessada {
		t.Fatal("primeira chamada nao deveria estar marcada como processada")
	}
	if len(resultado.Entrada.Itens) != 1 || resultado.Entrada.Itens[0].SaldoFisicoAtual != 15 {
		t.Fatalf("itens=%+v", resultado.Entrada.Itens)
	}
	var saldoFisico, custo float64
	if err = db.QueryRow(ctx, "SELECT saldo_fisico, custo_unitario FROM item_estoque WHERE id=$1", insumoID).Scan(&saldoFisico, &custo); err != nil || saldoFisico != 15 || custo != 25.00 {
		t.Fatalf("saldoFisico=%.2f custo=%.2f erro=%v", saldoFisico, custo, err)
	}
	var fornecedorMovimentacao string
	if err = db.QueryRow(ctx, "SELECT fornecedor_id FROM movimentacao_estoque WHERE item_estoque_id=$1 AND documento_origem=$2", insumoID, "NF-"+fmt.Sprintf("%012x", suffix)).Scan(&fornecedorMovimentacao); err != nil || fornecedorMovimentacao != fornecedorID {
		t.Fatalf("fornecedorMovimentacao=%q erro=%v", fornecedorMovimentacao, err)
	}
	var auditorias int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM auditoria_estoque WHERE item_estoque_id=$1 AND documento_origem=$2", insumoID, "NF-"+fmt.Sprintf("%012x", suffix)).Scan(&auditorias); err != nil || auditorias != 1 {
		t.Fatalf("auditorias=%d erro=%v", auditorias, err)
	}

	// Repetir a mesma Idempotency-Key: retorna a mesma resposta, sem duplicar saldo.
	repetido, err := useCase.Execute(ctx, application.RegistrarEntradaInput{
		IdempotencyKey: chaveSimples(suffix), DocumentoOrigem: "NF-" + fmt.Sprintf("%012x", suffix),
		Itens: []application.ItemInput{{ItemID: insumoID, Quantidade: 5, CustoUnitario: 25.00}},
	})
	if err != nil {
		t.Fatalf("entrada repetida: %v", err)
	}
	if !repetido.JaProcessada {
		t.Fatal("segunda chamada com a mesma chave deveria ser marcada como processada")
	}
	if err = db.QueryRow(ctx, "SELECT saldo_fisico FROM item_estoque WHERE id=$1", insumoID).Scan(&saldoFisico); err != nil || saldoFisico != 15 {
		t.Fatalf("saldoFisico apos repeticao=%.2f erro=%v", saldoFisico, err)
	}

	// Item inativo: rejeita sem alterar nada.
	if _, err = useCase.Execute(ctx, application.RegistrarEntradaInput{
		IdempotencyKey: "10000000-0000-0000-0000-000000000099", DocumentoOrigem: "NF-INATIVO",
		Itens: []application.ItemInput{{ItemID: pecaInativaID, Quantidade: 1, CustoUnitario: 10}},
	}); err != application.ErrItemInativo {
		t.Fatalf("erro=%v, esperado ErrItemInativo", err)
	}

	// Entrada vinculada ao pedido, cobrindo toda a necessidade: pedido conclui e a OS libera.
	comPedido, err := useCase.Execute(ctx, application.RegistrarEntradaInput{
		IdempotencyKey: chavePedido(suffix), DocumentoOrigem: "NF-PEDIDO-" + fmt.Sprintf("%012x", suffix),
		FornecedorID: fornecedorID, PedidoCompraID: pedidoID, Itens: []application.ItemInput{{ItemID: pecaPedidoID, Quantidade: 5, CustoUnitario: 12.00}},
	})
	if err != nil {
		t.Fatalf("entrada com pedido: %v", err)
	}
	if comPedido.Entrada.PedidoCompra == nil || comPedido.Entrada.PedidoCompra.Status != "CONCLUIDO" {
		t.Fatalf("pedidoCompra=%+v", comPedido.Entrada.PedidoCompra)
	}
	if len(comPedido.Entrada.OrdensServico) != 1 || comPedido.Entrada.OrdensServico[0].Status != "AGUARDANDO_EXECUCAO" {
		t.Fatalf("ordensServico=%+v", comPedido.Entrada.OrdensServico)
	}
	var statusOS string
	var dataEntradaFila time.Time
	var versaoOS int
	if err = db.QueryRow(ctx, "SELECT status, data_entrada_fila, version FROM ordem_servico WHERE id=$1", osID).Scan(&statusOS, &dataEntradaFila, &versaoOS); err != nil || statusOS != "AGUARDANDO_EXECUCAO" || dataEntradaFila.IsZero() || versaoOS != 2 {
		t.Fatalf("statusOS=%q dataEntradaFila=%v version=%d erro=%v", statusOS, dataEntradaFila, versaoOS, err)
	}
	if comPedido.Entrada.OrdensServico[0].DataEntradaFila == nil || comPedido.Entrada.OrdensServico[0].Version != versaoOS {
		t.Fatalf("ordemServico liberada=%+v", comPedido.Entrada.OrdensServico[0])
	}
	var saldoReservado int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM reserva_estoque WHERE ordem_servico_item_id=$1 AND status='ATIVA'", osItemID).Scan(&saldoReservado); err != nil || saldoReservado != 1 {
		t.Fatalf("reservas ativas=%d erro=%v", saldoReservado, err)
	}
	var quantidadeReservada float64
	if err = db.QueryRow(ctx, "SELECT quantidade_reservada FROM ordem_servico_item WHERE id=$1", osItemID).Scan(&quantidadeReservada); err != nil || quantidadeReservada != 5 {
		t.Fatalf("quantidadeReservada=%v erro=%v", quantidadeReservada, err)
	}
	if err = db.QueryRow(ctx, "SELECT quantidade_reservada FROM pedido_compra_item WHERE id=$1", pedidoItemID).Scan(&quantidadeReservada); err != nil || quantidadeReservada != 5 {
		t.Fatalf("quantidadeReservadaPedido=%v erro=%v", quantidadeReservada, err)
	}
}

func chaveSimples(suffix int64) string { return fmt.Sprintf("c1000000-0000-0000-0000-%012x", suffix) }
func chavePedido(suffix int64) string  { return fmt.Sprintf("c2000000-0000-0000-0000-%012x", suffix) }
