package ordemservico

import (
	"context"

	"github.com/jackc/pgx/v5"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

type devolucaoItemRow struct {
	osItemID            string
	itemEstoqueID       string
	quantidadeConsumida float64
	quantidadeReservada float64
	codigo, descricao   string
	tipo, unidadeMedida string
	saldoFisico         float64
	saldoReservado      float64
	ativo               bool
}

// DevolverItensAoEstoque libera reservas ativas e retorna ao saldo fisico o que ja foi consumido.
// Abre a propria transacao; para chamar dentro de uma transacao existente, use DevolverItensTx.

func (repository PostgresRepository) DevolverItensAoEstoque(ctx context.Context, ordemServicoID string) (domainEstoque.ResultadoDevolucao, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domainEstoque.ResultadoDevolucao{}, err
	}
	defer tx.Rollback(ctx)

	resultado, err := DevolverItensTx(ctx, tx, ordemServicoID, nil)
	if err != nil {
		return domainEstoque.ResultadoDevolucao{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domainEstoque.ResultadoDevolucao{}, err
	}
	return resultado, nil
}

// DevolverItensTx e a versao reentrante de DevolverItensAoEstoque, para ser chamada dentro da
// transacao de outro caso de uso (ex.: RecusarOrcamento). Quando itemEstoqueIDs e nao vazio,
// restringe a devolucao a esses itens (usado na recusa de um orcamento complementar); nil devolve
// todos os itens da OS (usado na recusa do orcamento principal, que cancela a OS inteira).

func DevolverItensTx(ctx context.Context, tx pgx.Tx, ordemServicoID string, itemEstoqueIDs []string) (domainEstoque.ResultadoDevolucao, error) {
	query := `
		SELECT osi.id, ie.id, osi.quantidade_consumida, osi.quantidade_reservada, ie.codigo, ie.descricao, ie.tipo, ie.unidade_medida,
		       ie.saldo_fisico, ie.saldo_reservado, ie.ativo
		FROM ordem_servico_item osi
		JOIN item_estoque ie ON ie.id = osi.item_estoque_id
		WHERE osi.ordem_servico_id = $1`
	args := []any{ordemServicoID}
	if len(itemEstoqueIDs) > 0 {
		query += " AND ie.id = ANY($2)"
		args = append(args, itemEstoqueIDs)
	}
	query += " ORDER BY ie.id FOR UPDATE OF ie"

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return domainEstoque.ResultadoDevolucao{}, err
	}
	var itens []devolucaoItemRow
	for rows.Next() {
		var item devolucaoItemRow
		if err = rows.Scan(&item.osItemID, &item.itemEstoqueID, &item.quantidadeConsumida, &item.quantidadeReservada, &item.codigo, &item.descricao,
			&item.tipo, &item.unidadeMedida, &item.saldoFisico, &item.saldoReservado, &item.ativo); err != nil {
			rows.Close()
			return domainEstoque.ResultadoDevolucao{}, err
		}
		itens = append(itens, item)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return domainEstoque.ResultadoDevolucao{}, err
	}

	resultado := domainEstoque.ResultadoDevolucao{OrdemServicoID: ordemServicoID}
	for _, item := range itens {
		processado := false

		quantidadeReservada, err := liberarReservasAtivas(ctx, tx, item.osItemID)
		if err != nil {
			return domainEstoque.ResultadoDevolucao{}, err
		}
		if quantidadeReservada > 0 {
			novoSaldoReservado := item.saldoReservado - quantidadeReservada
			novaQuantidadeReservada := item.quantidadeReservada - quantidadeReservada
			if novoSaldoReservado < 0 || novaQuantidadeReservada < 0 {
				return domainEstoque.ResultadoDevolucao{}, domainEstoque.ErrSaldoReservadoInsuficiente
			}
			if _, err = tx.Exec(ctx, "UPDATE item_estoque SET saldo_reservado = $2 WHERE id = $1", item.itemEstoqueID, novoSaldoReservado); err != nil {
				return domainEstoque.ResultadoDevolucao{}, err
			}
			if _, err = tx.Exec(ctx, "UPDATE ordem_servico_item SET quantidade_reservada = $2 WHERE id = $1", item.osItemID, novaQuantidadeReservada); err != nil {
				return domainEstoque.ResultadoDevolucao{}, err
			}
			if _, err = tx.Exec(ctx, `
				INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, tipo, quantidade)
				VALUES ($1, $2, $3, $4)`, item.itemEstoqueID, ordemServicoID, domainEstoque.MovimentacaoLiberacaoReserva, quantidadeReservada,
			); err != nil {
				return domainEstoque.ResultadoDevolucao{}, err
			}
			resultado.ReservasLiberadas = append(resultado.ReservasLiberadas, domainEstoque.ItemLiberado{
				ItemID: item.itemEstoqueID, Codigo: item.codigo, Descricao: item.descricao, Tipo: item.tipo,
				UnidadeMedida: item.unidadeMedida, Quantidade: quantidadeReservada, SaldoReservadoApos: novoSaldoReservado, Ativo: item.ativo,
			})
			processado = true
		}

		if item.quantidadeConsumida > 0 {
			novoSaldoFisico := item.saldoFisico + item.quantidadeConsumida
			if _, err = tx.Exec(ctx, "UPDATE item_estoque SET saldo_fisico = $2 WHERE id = $1", item.itemEstoqueID, novoSaldoFisico); err != nil {
				return domainEstoque.ResultadoDevolucao{}, err
			}
			if _, err = tx.Exec(ctx, "UPDATE ordem_servico_item SET quantidade_consumida = 0 WHERE id = $1", item.osItemID); err != nil {
				return domainEstoque.ResultadoDevolucao{}, err
			}
			if _, err = tx.Exec(ctx, `
				INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, tipo, quantidade)
				VALUES ($1, $2, $3, $4)`, item.itemEstoqueID, ordemServicoID, domainEstoque.MovimentacaoEntradaRetorno, item.quantidadeConsumida,
			); err != nil {
				return domainEstoque.ResultadoDevolucao{}, err
			}
			resultado.ItensRetornadosAoEstoque = append(resultado.ItensRetornadosAoEstoque, domainEstoque.ItemRetornado{
				ItemID: item.itemEstoqueID, Codigo: item.codigo, Descricao: item.descricao, Tipo: item.tipo,
				UnidadeMedida: item.unidadeMedida, Quantidade: item.quantidadeConsumida, SaldoFisicoApos: novoSaldoFisico, Ativo: item.ativo,
			})
			processado = true
		}

		pendentes, err := desvincularPedidosPendentes(ctx, tx, item.osItemID)
		if err != nil {
			return domainEstoque.ResultadoDevolucao{}, err
		}
		for _, pendente := range pendentes {
			resultado.ItensSemDevolucao = append(resultado.ItensSemDevolucao, domainEstoque.ItemSemDevolucao{
				ItemID: item.itemEstoqueID, Codigo: item.codigo, Descricao: item.descricao, Tipo: item.tipo,
				UnidadeMedida: item.unidadeMedida, Quantidade: pendente.quantidade,
				Motivo: domainEstoque.MotivoPedidoDeCompraNaoRecebido, PedidoID: pendente.pedidoID, Ativo: item.ativo,
			})
			processado = true
		}

		if processado {
			resultado.TotalItensProcessados++
		}
	}

	return resultado, nil
}

func liberarReservasAtivas(ctx context.Context, tx pgx.Tx, osItemID string) (float64, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, quantidade FROM reserva_estoque
		WHERE ordem_servico_item_id = $1 AND status = $2
		FOR UPDATE`, osItemID, domainEstoque.ReservaAtiva)
	if err != nil {
		return 0, err
	}
	type reserva struct {
		id         string
		quantidade float64
	}
	var reservas []reserva
	for rows.Next() {
		var r reserva
		if err = rows.Scan(&r.id, &r.quantidade); err != nil {
			rows.Close()
			return 0, err
		}
		reservas = append(reservas, r)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return 0, err
	}

	var total float64
	for _, r := range reservas {
		if _, err = tx.Exec(ctx, `
			UPDATE reserva_estoque SET status = $2, liberada_em = CURRENT_TIMESTAMP WHERE id = $1`,
			r.id, domainEstoque.ReservaLiberada,
		); err != nil {
			return 0, err
		}
		total += r.quantidade
	}
	return total, nil
}

type pedidoPendente struct {
	pedidoID   string
	quantidade float64
}

func desvincularPedidosPendentes(ctx context.Context, tx pgx.Tx, osItemID string) ([]pedidoPendente, error) {
	rows, err := tx.Query(ctx, `
		SELECT pci.pedido_compra_id, pcio.quantidade_atendida
		FROM pedido_compra_item_os pcio
		JOIN pedido_compra_item pci ON pci.id = pcio.pedido_compra_item_id
		JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
		WHERE pcio.ordem_servico_item_id = $1 AND pc.status <> 'CONCLUIDO'`, osItemID)
	if err != nil {
		return nil, err
	}
	var pendentes []pedidoPendente
	for rows.Next() {
		var p pedidoPendente
		if err = rows.Scan(&p.pedidoID, &p.quantidade); err != nil {
			rows.Close()
			return nil, err
		}
		pendentes = append(pendentes, p)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `
		DELETE FROM pedido_compra_item_os
		WHERE ordem_servico_item_id = $1
		  AND pedido_compra_item_id IN (
		      SELECT pci.id FROM pedido_compra_item pci
		      JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
		      WHERE pc.status <> 'CONCLUIDO'
		  )`, osItemID); err != nil {
		return nil, err
	}
	return pendentes, nil
}
