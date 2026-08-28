package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
)

func TestDevolverItensAoEstoque(t *testing.T) {
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
	clienteID, veiculoID, categoriaID := id("a1000000-0000-0000-0000-"), id("a2000000-0000-0000-0000-"), id("a3000000-0000-0000-0000-")
	osID := id("a4000000-0000-0000-0000-")
	pecaReservadaID, insumoConsumidoID, pecaPendenteID := id("a5000000-0000-0000-0000-"), id("a6000000-0000-0000-0000-"), id("a7000000-0000-0000-0000-")
	osItemReservadoID, osItemConsumidoID, osItemPendenteID := id("a8000000-0000-0000-0000-"), id("a9000000-0000-0000-0000-"), id("aa000000-0000-0000-0000-")
	reservaAtivaID := id("ab000000-0000-0000-0000-")
	fornecedorID, pedidoCompraID, pedidoCompraItemID := id("ac000000-0000-0000-0000-"), id("ad000000-0000-0000-0000-"), id("ae000000-0000-0000-0000-")
	placa := fmt.Sprintf("DEV1A%02d", suffix%100)
	codigoBase := suffix & 0xffffff

	if _, err = db.Exec(ctx, "INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)", categoriaID, fmt.Sprintf("Categoria %x", codigoBase)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES ($1,$2,$3,$4,'CANCELADA')", osID, clienteID, veiculoID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO item_estoque (id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado) VALUES
		($1,$4,'PECA',$5,'Peca reservada','Peca reservada','peca reservada','UN',10,3),
		($2,$4,'INSUMO',$6,'Insumo consumido','Insumo consumido','insumo consumido','L',20,2),
		($3,$4,'PECA',$7,'Peca pendente','Peca pendente','peca pendente','UN',0,0)`,
		pecaReservadaID, insumoConsumidoID, pecaPendenteID, categoriaID,
		fmt.Sprintf("PEC%06x1", codigoBase), fmt.Sprintf("INS%06x1", codigoBase), fmt.Sprintf("PEC%06x2", codigoBase)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO ordem_servico_item (id,ordem_servico_id,item_estoque_id,quantidade_necessaria,quantidade_reservada,quantidade_consumida,valor_unitario) VALUES
		($1,$4,$5,3,3,0,10),
		($2,$4,$6,2,0,2,5),
		($3,$4,$7,1,0,0,8)`,
		osItemReservadoID, osItemConsumidoID, osItemPendenteID, osID, pecaReservadaID, insumoConsumidoID, pecaPendenteID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO reserva_estoque (id,ordem_servico_item_id,item_estoque_id,quantidade,status) VALUES ($1,$2,$3,3,'ATIVA')`,
		reservaAtivaID, osItemReservadoID, pecaReservadaID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO fornecedor (id,razao_social,documento,tipo_documento) VALUES ($1,'Fornecedor Teste',$2,'CNPJ')",
		fornecedorID, fmt.Sprintf("%014d", suffix%100000000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO pedido_compra (id,fornecedor_id,numero,status) VALUES ($1,$2,$3,'ABERTO')",
		pedidoCompraID, fornecedorID, fmt.Sprintf("DEV/%012x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO pedido_compra_item (id,pedido_compra_id,item_estoque_id,quantidade_necessaria,quantidade_pedida) VALUES ($1,$2,$3,1,1)`,
		pedidoCompraItemID, pedidoCompraID, pecaPendenteID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO pedido_compra_item_os (pedido_compra_item_id,ordem_servico_item_id,quantidade_atendida) VALUES ($1,$2,1)",
		pedidoCompraItemID, osItemPendenteID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM movimentacao_estoque WHERE ordem_servico_id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra_item_os WHERE pedido_compra_item_id=$1", pedidoCompraItemID)
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra_item WHERE id=$1", pedidoCompraItemID)
		_, _ = db.Exec(ctx, "DELETE FROM pedido_compra WHERE id=$1", pedidoCompraID)
		_, _ = db.Exec(ctx, "DELETE FROM fornecedor WHERE id=$1", fornecedorID)
		_, _ = db.Exec(ctx, "DELETE FROM reserva_estoque WHERE ordem_servico_item_id IN ($1,$2,$3)", osItemReservadoID, osItemConsumidoID, osItemPendenteID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE ordem_servico_id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id IN ($1,$2,$3)", pecaReservadaID, insumoConsumidoID, pecaPendenteID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id=$1", categoriaID)
	})

	useCase := application.NewDevolverItensAoEstoque(infrastructure.NewPostgresRepository(db))
	resultado, err := useCase.Execute(ctx, osID)
	if err != nil {
		t.Fatalf("devolucao: %v", err)
	}
	if resultado.TotalItensProcessados != 3 {
		t.Fatalf("totalItensProcessados=%d, esperado 3", resultado.TotalItensProcessados)
	}
	if len(resultado.ReservasLiberadas) != 1 || resultado.ReservasLiberadas[0].Quantidade != 3 || resultado.ReservasLiberadas[0].SaldoReservadoApos != 0 {
		t.Fatalf("reservasLiberadas=%+v", resultado.ReservasLiberadas)
	}
	if len(resultado.ItensRetornadosAoEstoque) != 1 || resultado.ItensRetornadosAoEstoque[0].Quantidade != 2 || resultado.ItensRetornadosAoEstoque[0].SaldoFisicoApos != 22 {
		t.Fatalf("itensRetornados=%+v", resultado.ItensRetornadosAoEstoque)
	}
	if len(resultado.ItensSemDevolucao) != 1 || resultado.ItensSemDevolucao[0].Motivo != "PEDIDO_DE_COMPRA_NAO_RECEBIDO" || resultado.ItensSemDevolucao[0].PedidoID != pedidoCompraID {
		t.Fatalf("itensSemDevolucao=%+v", resultado.ItensSemDevolucao)
	}

	var statusReserva string
	if err = db.QueryRow(ctx, "SELECT status FROM reserva_estoque WHERE id=$1", reservaAtivaID).Scan(&statusReserva); err != nil || statusReserva != "LIBERADA" {
		t.Fatalf("status reserva=%q erro=%v", statusReserva, err)
	}
	var saldoReservado, saldoFisico float64
	if err = db.QueryRow(ctx, "SELECT saldo_reservado FROM item_estoque WHERE id=$1", pecaReservadaID).Scan(&saldoReservado); err != nil || saldoReservado != 0 {
		t.Fatalf("saldoReservado=%.2f erro=%v", saldoReservado, err)
	}
	if err = db.QueryRow(ctx, "SELECT saldo_fisico FROM item_estoque WHERE id=$1", insumoConsumidoID).Scan(&saldoFisico); err != nil || saldoFisico != 22 {
		t.Fatalf("saldoFisico=%.2f erro=%v", saldoFisico, err)
	}
	var quantidadeConsumida float64
	if err = db.QueryRow(ctx, "SELECT quantidade_consumida FROM ordem_servico_item WHERE id=$1", osItemConsumidoID).Scan(&quantidadeConsumida); err != nil || quantidadeConsumida != 0 {
		t.Fatalf("quantidadeConsumida=%.2f erro=%v", quantidadeConsumida, err)
	}
	var vinculosPendentes int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM pedido_compra_item_os WHERE ordem_servico_item_id=$1", osItemPendenteID).Scan(&vinculosPendentes); err != nil || vinculosPendentes != 0 {
		t.Fatalf("vinculosPendentes=%d erro=%v", vinculosPendentes, err)
	}
	var movimentacoes int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM movimentacao_estoque WHERE ordem_servico_id=$1", osID).Scan(&movimentacoes); err != nil || movimentacoes != 2 {
		t.Fatalf("movimentacoes=%d erro=%v", movimentacoes, err)
	}

	repetido, err := useCase.Execute(ctx, osID)
	if err != nil {
		t.Fatalf("devolucao repetida: %v", err)
	}
	if repetido.TotalItensProcessados != 0 {
		t.Fatalf("devolucao repetida processou %d itens, esperado 0 (idempotencia)", repetido.TotalItensProcessados)
	}
}
