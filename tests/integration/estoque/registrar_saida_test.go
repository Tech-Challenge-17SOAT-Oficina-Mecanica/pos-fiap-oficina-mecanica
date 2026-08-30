package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/estoque"
)

func TestRegistrarSaidaEstoque(t *testing.T) {
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
	categoriaID := id("d1000000-0000-0000-0000-")
	itemID := id("d2000000-0000-0000-0000-")
	clienteID, veiculoID := id("d3000000-0000-0000-0000-"), id("d4000000-0000-0000-0000-")
	osID, osItemID, reservaID := id("d5000000-0000-0000-0000-"), id("d6000000-0000-0000-0000-"), id("d7000000-0000-0000-0000-")
	chave := fmt.Sprintf("d8000000-0000-0000-0000-%012x", suffix)
	placa := fmt.Sprintf("S%06X", suffix&0xffffff)
	codigo := fmt.Sprintf("P%06X", suffix&0xffffff)

	if _, err = db.Exec(ctx, "INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)", categoriaID, fmt.Sprintf("Categoria saida %x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO item_estoque
		(id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado,ativo,custo_unitario)
		VALUES ($1,$2,'PECA',$3,'Peca saida','Peca saida','peca saida','UN',10,5,true,12.50)`,
		itemID, categoriaID, codigo); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Saida',$2,'CPF','11999999999')",
		clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Saida',2024)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES ($1,$2,$3,$4,'EM_EXECUCAO')", osID, clienteID, veiculoID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico_item (id,ordem_servico_id,item_estoque_id,quantidade_necessaria,quantidade_reservada,valor_unitario) VALUES ($1,$2,$3,5,5,20.00)", osItemID, osID, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO reserva_estoque (id,ordem_servico_item_id,item_estoque_id,quantidade,status) VALUES ($1,$2,$3,5,$4)", reservaID, osItemID, itemID, domain.ReservaAtiva); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_estoque WHERE item_estoque_id=$1", itemID)
		_, _ = db.Exec(ctx, "DELETE FROM movimentacao_estoque WHERE item_estoque_id=$1", itemID)
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM chave_idempotencia WHERE chave=$1", chave)
		_, _ = db.Exec(ctx, "DELETE FROM reserva_estoque WHERE ordem_servico_item_id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id=$1", itemID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id=$1", categoriaID)
	})

	useCase := application.NewRegistrarSaida(infrastructure.NewPostgresRepository(db))
	resultado, err := useCase.Execute(ctx, application.RegistrarSaidaInput{
		IdempotencyKey: chave, OrdemServicoID: osID, LiberarSaldoNaoUsado: true,
		Itens: []application.ItemSaidaInput{{ItemID: itemID, Quantidade: 4}},
	})
	if err != nil {
		t.Fatalf("registrar saida: %v", err)
	}
	if resultado.JaProcessada || resultado.Saida.CustoTotalSaida != 50 || len(resultado.Saida.Itens) != 1 {
		t.Fatalf("resultado=%+v", resultado)
	}
	item := resultado.Saida.Itens[0]
	if item.QuantidadeReservadaAntes != 5 || item.QuantidadeLiberada != 1 || item.SaldoFisicoAtual != 6 || item.SaldoReservadoAtual != 0 {
		t.Fatalf("item=%+v", item)
	}

	var saldoFisico, saldoReservado, quantidadeReservada, quantidadeConsumida, custoMateriais float64
	if err = db.QueryRow(ctx, "SELECT saldo_fisico, saldo_reservado FROM item_estoque WHERE id=$1", itemID).Scan(&saldoFisico, &saldoReservado); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(ctx, "SELECT quantidade_reservada, quantidade_consumida FROM ordem_servico_item WHERE id=$1", osItemID).Scan(&quantidadeReservada, &quantidadeConsumida); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(ctx, "SELECT custo_total_materiais FROM ordem_servico WHERE id=$1", osID).Scan(&custoMateriais); err != nil {
		t.Fatal(err)
	}
	if saldoFisico != 6 || saldoReservado != 0 || quantidadeReservada != 0 || quantidadeConsumida != 4 || custoMateriais != 50 {
		t.Fatalf("saldos fisico=%.2f reservado=%.2f osReservado=%.2f consumido=%.2f custo=%.2f", saldoFisico, saldoReservado, quantidadeReservada, quantidadeConsumida, custoMateriais)
	}
	var reservasAtivas int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM reserva_estoque WHERE ordem_servico_item_id=$1 AND status=$2", osItemID, domain.ReservaAtiva).Scan(&reservasAtivas); err != nil || reservasAtivas != 0 {
		t.Fatalf("reservasAtivas=%d erro=%v", reservasAtivas, err)
	}

	repetido, err := useCase.Execute(ctx, application.RegistrarSaidaInput{
		IdempotencyKey: chave, OrdemServicoID: osID, LiberarSaldoNaoUsado: true,
		Itens: []application.ItemSaidaInput{{ItemID: itemID, Quantidade: 4}},
	})
	if err != nil {
		t.Fatalf("registrar saida repetida: %v", err)
	}
	if !repetido.JaProcessada {
		t.Fatal("segunda chamada deveria ser idempotente")
	}
	if err = db.QueryRow(ctx, "SELECT saldo_fisico, saldo_reservado FROM item_estoque WHERE id=$1", itemID).Scan(&saldoFisico, &saldoReservado); err != nil || saldoFisico != 6 || saldoReservado != 0 {
		t.Fatalf("saldos apos repeticao fisico=%.2f reservado=%.2f erro=%v", saldoFisico, saldoReservado, err)
	}
}
