package ordemservico

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

func (repository PostgresRepository) IniciarExecucao(ctx context.Context, input application.IniciarExecucaoInput) (domain.ResultadoInicioExecucao, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}
	defer tx.Rollback(ctx)

	var status, mecanicoResponsavelID, clienteID string
	err = tx.QueryRow(ctx, `
		SELECT status, COALESCE(mecanico_responsavel_id::text, ''), cliente_id::text
		FROM ordem_servico WHERE id = $1 FOR UPDATE`, input.OSID,
	).Scan(&status, &mecanicoResponsavelID, &clienteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ResultadoInicioExecucao{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}

	ordem := domain.OrdemDeServico{ID: input.OSID, Status: status}
	if err = ordem.IniciarExecucao(time.Now()); err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}

	var possuiOrcamentoAprovado, possuiServicosAutorizados bool
	if err = tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM orcamento WHERE ordem_servico_id = $1 AND status = 'APROVADO'),
		       EXISTS (
			   SELECT 1 FROM orcamento o
			   JOIN orcamento_item oi ON oi.orcamento_id = o.id
			   WHERE o.ordem_servico_id = $1 AND o.status = 'APROVADO' AND oi.tipo_item = 'SERVICO'
		       )`, input.OSID,
	).Scan(&possuiOrcamentoAprovado, &possuiServicosAutorizados); err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}
	if !possuiOrcamentoAprovado {
		return domain.ResultadoInicioExecucao{}, domain.ErrOrcamentoNaoAprovado
	}
	if !possuiServicosAutorizados {
		return domain.ResultadoInicioExecucao{}, domain.ErrServicosNaoAutorizados
	}

	var possuiRecursoSemReserva bool
	if err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM ordem_servico_item osi
			WHERE osi.ordem_servico_id = $1
			  AND osi.quantidade_necessaria > COALESCE((
				SELECT SUM(r.quantidade) FROM reserva_estoque r
				WHERE r.ordem_servico_item_id = osi.id AND r.status = 'ATIVA'
			), 0)
		)`, input.OSID,
	).Scan(&possuiRecursoSemReserva); err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}
	if possuiRecursoSemReserva {
		return domain.ResultadoInicioExecucao{}, domain.ErrRecursosIndisponiveis
	}

	var mecanicoAutenticadoID string
	if err = tx.QueryRow(ctx, "SELECT id FROM mecanico WHERE usuario_id = $1", input.UsuarioID).Scan(&mecanicoAutenticadoID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ResultadoInicioExecucao{}, domain.ErrMecanicoAutenticadoNaoEncontrado
		}
		return domain.ResultadoInicioExecucao{}, err
	}
	if mecanicoResponsavelID == "" {
		mecanicoResponsavelID = mecanicoAutenticadoID
	}

	itensBaixados, custoTotalMateriais, err := baixarReservasNoInicioDaExecucao(ctx, tx, input.OSID, input.UsuarioID)
	if err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}

	var dataInicio time.Time
	if err = tx.QueryRow(ctx, `
		UPDATE ordem_servico
		SET status = $2, iniciada_em = CURRENT_TIMESTAMP,
		    custo_total_materiais = custo_total_materiais + $4::numeric,
		    mecanico_responsavel_id = NULLIF($3, '')::uuid
		WHERE id = $1
		RETURNING iniciada_em`, input.OSID, ordem.Status, mecanicoResponsavelID, custoTotalMateriais,
	).Scan(&dataInicio); err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico
		(ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em)
		VALUES ($1, NULLIF($2, '')::uuid, 'ORDEM_SERVICO', $1, 'EXECUCAO_INICIADA',
		        jsonb_build_object('status', $3::text, 'mecanicoId', $4::text), '{}'::jsonb, $5)`,
		input.OSID, input.UsuarioID, ordem.Status, mecanicoResponsavelID, dataInicio,
	); err != nil {
		return domain.ResultadoInicioExecucao{}, fmt.Errorf("registrar auditoria do inicio da execucao: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}
	return domain.ResultadoInicioExecucao{
		OrdemServicoID: input.OSID, ClienteID: clienteID, Status: ordem.Status, MecanicoID: mecanicoResponsavelID, DataInicioExecucao: dataInicio,
		ItensBaixados: itensBaixados, CustoTotalMateriaisBaixados: custoTotalMateriais,
	}, nil
}

type reservaInicioExecucaoRow struct {
	reservaID, osItemID, itemID string
	codigo, tipo, unidadeMedida string
	quantidade, custoUnitario   float64
}

func baixarReservasNoInicioDaExecucao(ctx context.Context, tx pgx.Tx, ordemServicoID, usuarioID string) ([]domain.ItemBaixadoInicioExecucao, float64, error) {
	rows, err := tx.Query(ctx, `
		SELECT r.id, osi.id, ie.id, ie.codigo, ie.tipo, ie.unidade_medida, r.quantidade, COALESCE(ie.custo_unitario, 0)
		FROM reserva_estoque r
		JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id
		JOIN item_estoque ie ON ie.id = r.item_estoque_id
		WHERE osi.ordem_servico_id = $1 AND r.status = $2
		ORDER BY ie.id, r.id
		FOR UPDATE OF r, osi, ie`, ordemServicoID, domainEstoque.ReservaAtiva)
	if err != nil {
		return nil, 0, err
	}
	// O cursor precisa ser consumido e fechado ANTES de qualquer outra query: o pgx
	// mantem a conexao ocupada enquanto ha Rows aberto, e escrever dentro do loop
	// devolve "conn busy".
	var reservas []reservaInicioExecucaoRow
	for rows.Next() {
		var reserva reservaInicioExecucaoRow
		if err = rows.Scan(&reserva.reservaID, &reserva.osItemID, &reserva.itemID, &reserva.codigo, &reserva.tipo, &reserva.unidadeMedida, &reserva.quantidade, &reserva.custoUnitario); err != nil {
			rows.Close()
			return nil, 0, err
		}
		reservas = append(reservas, reserva)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	porItem := map[string]int{}
	var itens []domain.ItemBaixadoInicioExecucao
	var custoTotal float64
	for _, reserva := range reservas {
		custo := reserva.quantidade * reserva.custoUnitario
		var saldoFisicoAtual, saldoReservadoAtual float64
		if err = tx.QueryRow(ctx, `
			UPDATE item_estoque
			SET saldo_fisico = saldo_fisico - $2::numeric,
			    saldo_reservado = saldo_reservado - $2::numeric
			WHERE id = $1
			  AND saldo_fisico >= $2::numeric
			  AND saldo_reservado >= $2::numeric
			RETURNING saldo_fisico, saldo_reservado`, reserva.itemID, reserva.quantidade,
		).Scan(&saldoFisicoAtual, &saldoReservadoAtual); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, 0, domain.ErrRecursosIndisponiveis
			}
			return nil, 0, err
		}
		if _, err = tx.Exec(ctx, `
			UPDATE ordem_servico_item
			SET quantidade_reservada = quantidade_reservada - $2::numeric,
			    quantidade_consumida = quantidade_consumida + $2::numeric
			WHERE id = $1`, reserva.osItemID, reserva.quantidade); err != nil {
			return nil, 0, err
		}
		if _, err = tx.Exec(ctx, "UPDATE reserva_estoque SET status = $2, liberada_em = CURRENT_TIMESTAMP WHERE id = $1", reserva.reservaID, domainEstoque.ReservaConsumida); err != nil {
			return nil, 0, err
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, reserva_estoque_id, tipo, quantidade, custo_unitario)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			reserva.itemID, ordemServicoID, reserva.reservaID, domainEstoque.MovimentacaoSaida, reserva.quantidade, reserva.custoUnitario,
		); err != nil {
			return nil, 0, err
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO auditoria_estoque (item_estoque_id, usuario_id, tipo_evento, documento_origem, dados, ocorrido_em)
			VALUES ($1, NULLIF($2, '')::uuid, 'SAIDA_ESTOQUE', $3,
			        jsonb_build_object('ordemServicoId', $4::text, 'reservaId', $5::text, 'quantidadeBaixada', $6::numeric),
			        CURRENT_TIMESTAMP)`,
			reserva.itemID, usuarioID, "INICIO-"+ordemServicoID, ordemServicoID, reserva.reservaID, reserva.quantidade,
		); err != nil {
			return nil, 0, err
		}

		if index, existe := porItem[reserva.itemID]; existe {
			itens[index].QuantidadeBaixada += reserva.quantidade
			itens[index].SaldoFisicoAtual = saldoFisicoAtual
			itens[index].SaldoReservadoAtual = saldoReservadoAtual
			itens[index].CustoTotal += custo
		} else {
			porItem[reserva.itemID] = len(itens)
			itens = append(itens, domain.ItemBaixadoInicioExecucao{
				ItemID: reserva.itemID, Codigo: reserva.codigo, Tipo: reserva.tipo, UnidadeMedida: reserva.unidadeMedida,
				QuantidadeBaixada: reserva.quantidade, SaldoFisicoAtual: saldoFisicoAtual, SaldoReservadoAtual: saldoReservadoAtual,
				CustoUnitario: reserva.custoUnitario, CustoTotal: custo,
			})
		}
		custoTotal += custo
	}
	return itens, custoTotal, nil
}

func (repository PostgresRepository) Entregar(ctx context.Context, input application.EntregarInput) (domain.ResultadoEntrega, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.ResultadoEntrega{}, err
	}
	defer tx.Rollback(ctx)

	var status, clienteOriginalID string
	if err = tx.QueryRow(ctx, `SELECT status, cliente_id FROM ordem_servico WHERE id = $1 FOR UPDATE`, input.OSID).
		Scan(&status, &clienteOriginalID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ResultadoEntrega{}, application.ErrOrdemServicoNaoEncontrada
		}
		return domain.ResultadoEntrega{}, err
	}
	if err = domain.ValidarEntrega(status); err != nil {
		return domain.ResultadoEntrega{}, err
	}

	clienteRetiradaID := input.ClienteID
	if clienteRetiradaID == "" {
		clienteRetiradaID = clienteOriginalID
	} else {
		var ativo bool
		if err = tx.QueryRow(ctx, `SELECT ativo FROM cliente WHERE id = $1`, clienteRetiradaID).Scan(&ativo); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ResultadoEntrega{}, application.ErrClienteNaoEncontrado
			}
			return domain.ResultadoEntrega{}, err
		}
		if !ativo {
			return domain.ResultadoEntrega{}, application.ErrClienteNaoEncontrado
		}
	}

	var quantidadeOrcamentos int
	var valorFinal float64
	if err = tx.QueryRow(ctx, `
		SELECT COUNT(DISTINCT o.id), COALESCE(SUM(oi.valor_total), 0)
		FROM orcamento o
		LEFT JOIN orcamento_item oi ON oi.orcamento_id = o.id
		WHERE o.ordem_servico_id = $1 AND o.status = 'APROVADO'`, input.OSID,
	).Scan(&quantidadeOrcamentos, &valorFinal); err != nil {
		return domain.ResultadoEntrega{}, err
	}
	if quantidadeOrcamentos == 0 {
		return domain.ResultadoEntrega{}, domain.ErrValorFinalIndisponivel
	}

	var dataEntrega time.Time
	if err = tx.QueryRow(ctx, `
		UPDATE ordem_servico
		SET status = $2, valor_final = $3, entregue_em = CURRENT_TIMESTAMP,
		    responsavel_entrega_id = NULLIF($4, '')::uuid,
		    cliente_retirada_id = $5, observacoes_entrega = NULLIF($6, '')
		WHERE id = $1
		RETURNING entregue_em`, input.OSID, domain.StatusEntregue, valorFinal,
		input.UsuarioID, clienteRetiradaID, input.Observacoes,
	).Scan(&dataEntrega); err != nil {
		return domain.ResultadoEntrega{}, err
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico
		(ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em)
		VALUES ($1, NULLIF($2, '')::uuid, 'ORDEM_SERVICO', $1, 'VEICULO_ENTREGUE',
		        jsonb_build_object('status', $3::text, 'valorFinal', $4::numeric,
		                           'clienteRetiradaId', $5::text, 'observacoes', COALESCE($6, '')),
		        '{}'::jsonb, $7)`, input.OSID, input.UsuarioID, domain.StatusEntregue,
		valorFinal, clienteRetiradaID, input.Observacoes, dataEntrega,
	); err != nil {
		return domain.ResultadoEntrega{}, fmt.Errorf("registrar auditoria da entrega: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.ResultadoEntrega{}, err
	}
	return domain.ResultadoEntrega{
		OrdemServicoID: input.OSID, Status: domain.StatusEntregue, ValorFinal: valorFinal,
		ResponsavelEntregaID: input.UsuarioID, ClienteID: clienteRetiradaID,
		DataEntrega: dataEntrega, Observacoes: input.Observacoes,
	}, nil
}

// Listar aplica os filtros de GET /ordens-servico. Documento ou placa sem cadastro correspondente
// retornam erro (404), diferente de filtro sem resultado, que devolve lista vazia (200).

func (repository PostgresRepository) Finalizar(ctx context.Context, input application.FinalizarInput) (domain.ResultadoFinalizacao, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.ResultadoFinalizacao{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", input.OSID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ResultadoFinalizacao{}, application.ErrOrdemServicoNaoEncontrada
		}
		return domain.ResultadoFinalizacao{}, err
	}
	if status != domain.StatusEmExecucao {
		return domain.ResultadoFinalizacao{}, domain.ErrOSNaoEmExecucao
	}

	var servicosPendentes int
	if err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM ordem_servico_servico WHERE ordem_servico_id = $1 AND status <> 'CONCLUIDO'", input.OSID).Scan(&servicosPendentes); err != nil {
		return domain.ResultadoFinalizacao{}, err
	}
	if servicosPendentes > 0 {
		return domain.ResultadoFinalizacao{}, domain.ErrServicosPendentes
	}

	var complementarPendente int
	if err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = 'COMPLEMENTAR' AND status = 'CRIADO'", input.OSID).Scan(&complementarPendente); err != nil {
		return domain.ResultadoFinalizacao{}, err
	}
	if complementarPendente > 0 {
		return domain.ResultadoFinalizacao{}, domain.ErrOrcamentoComplementarPendente
	}

	itensPendentes, err := reservasAtivasDaOS(ctx, tx, input.OSID)
	if err != nil {
		return domain.ResultadoFinalizacao{}, err
	}
	if len(itensPendentes) > 0 {
		return domain.ResultadoFinalizacao{}, domain.ErroReservasPendentes{Itens: itensPendentes}
	}

	var dataFinalizacao time.Time
	var clienteID string
	if err = tx.QueryRow(ctx, `
		UPDATE ordem_servico SET status = $2, finalizada_em = CURRENT_TIMESTAMP, observacoes_finalizacao = NULLIF($3, '')
		WHERE id = $1
		RETURNING finalizada_em, cliente_id`, input.OSID, domain.StatusFinalizada, input.Observacoes,
	).Scan(&dataFinalizacao, &clienteID); err != nil {
		return domain.ResultadoFinalizacao{}, err
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em)
		VALUES ($1, NULLIF($2, '')::uuid, 'ORDEM_SERVICO', $1, 'FINALIZACAO', jsonb_build_object('observacoes', COALESCE($3, '')), '{}'::jsonb, $4)`,
		input.OSID, input.UsuarioID, input.Observacoes, dataFinalizacao,
	); err != nil {
		return domain.ResultadoFinalizacao{}, fmt.Errorf("registrar auditoria da finalizacao: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.ResultadoFinalizacao{}, err
	}

	// A notificacao nao acontece aqui: quem finaliza so devolve o cliente, e o caso de
	// uso enfileira o aviso depois do commit (RNF-OS-44).
	resultado := domain.ResultadoFinalizacao{
		OrdemServicoID: input.OSID, ClienteID: clienteID, Status: domain.StatusFinalizada,
		DataFinalizacao: dataFinalizacao, Observacoes: input.Observacoes,
	}
	return resultado, nil
}

func reservasAtivasDaOS(ctx context.Context, tx pgx.Tx, ordemServicoID string) ([]domain.ItemPendenteBaixa, error) {
	rows, err := tx.Query(ctx, `
		SELECT ie.id, ie.codigo, r.quantidade
		FROM reserva_estoque r
		JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id
		JOIN item_estoque ie ON ie.id = r.item_estoque_id
		WHERE osi.ordem_servico_id = $1 AND r.status = $2
		ORDER BY ie.id`, ordemServicoID, domainEstoque.ReservaAtiva)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var itens []domain.ItemPendenteBaixa
	for rows.Next() {
		var item domain.ItemPendenteBaixa
		if err = rows.Scan(&item.ItemID, &item.Codigo, &item.Quantidade); err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}
	return itens, rows.Err()
}
