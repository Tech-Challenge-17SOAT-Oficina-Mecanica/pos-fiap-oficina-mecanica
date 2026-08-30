package ordemservico

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domainOrcamento "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) ConsultarFila(ctx context.Context, limite, deslocamento int) ([]domain.ItemFila, int, error) {
	const filtro = `
		FROM ordem_servico os
		JOIN veiculo v ON v.id = os.veiculo_id
		WHERE os.status = 'AGUARDANDO_EXECUCAO'
		  AND os.data_entrada_fila IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1
			FROM ordem_servico_item osi
			JOIN item_estoque ie ON ie.id = osi.item_estoque_id
			WHERE osi.ordem_servico_id = os.id
			  AND osi.quantidade_reservada
			      + GREATEST(ie.saldo_fisico - ie.saldo_reservado, 0) < osi.quantidade_necessaria
		  )`
	var total int
	if err := repository.db.QueryRow(ctx, "SELECT COUNT(*) "+filtro).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.db.Query(ctx, `
		SELECT os.id, v.placa, v.marca, v.modelo, os.status,
		       os.mecanico_responsavel_id, os.data_entrada_fila`+filtro+`
		ORDER BY (os.mecanico_responsavel_id IS NULL), os.data_entrada_fila, os.id
		LIMIT $1 OFFSET $2`, limite, deslocamento)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	itens := []domain.ItemFila{}
	for rows.Next() {
		var item domain.ItemFila
		if err = rows.Scan(&item.OrdemServicoID, &item.Placa, &item.Marca, &item.Modelo,
			&item.Status, &item.MecanicoResponsavelID, &item.DataEntradaFila); err != nil {
			return nil, 0, err
		}
		itens = append(itens, item)
	}
	return itens, total, rows.Err()
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
func (repository PostgresRepository) Listar(ctx context.Context, filtros domain.FiltrosListagem, limite, deslocamento int) ([]domain.ItemListagem, int, error) {
	if filtros.Documento != "" {
		var existe bool
		if err := repository.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM cliente WHERE documento = $1)", filtros.Documento).Scan(&existe); err != nil {
			return nil, 0, err
		}
		if !existe {
			return nil, 0, application.ErrClienteNaoEncontrado
		}
	}
	if filtros.Placa != "" {
		var existe bool
		if err := repository.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM veiculo WHERE placa = $1)", filtros.Placa).Scan(&existe); err != nil {
			return nil, 0, err
		}
		if !existe {
			return nil, 0, application.ErrVeiculoNaoEncontrado
		}
	}

	const filtro = `
		FROM ordem_servico os
		JOIN cliente c ON c.id = os.cliente_id
		JOIN veiculo v ON v.id = os.veiculo_id
		WHERE ($1 = '' OR os.status = $1)
		  AND ($2 = '' OR c.documento = $2)
		  AND ($3 = '' OR v.placa = $3)`

	var total int
	if err := repository.db.QueryRow(ctx, "SELECT COUNT(*) "+filtro, filtros.Status, filtros.Documento, filtros.Placa).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := repository.db.Query(ctx, `
		SELECT os.id, os.status, c.id, c.nome, c.documento, v.id, v.placa, v.marca, v.modelo`+filtro+`
		ORDER BY os.criada_em DESC, os.id DESC
		LIMIT $4 OFFSET $5`, filtros.Status, filtros.Documento, filtros.Placa, limite, deslocamento)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	itens := []domain.ItemListagem{}
	for rows.Next() {
		var item domain.ItemListagem
		if err = rows.Scan(&item.OrdemServicoID, &item.Status, &item.ClienteID, &item.ClienteNome, &item.ClienteDocumento,
			&item.VeiculoID, &item.Placa, &item.Marca, &item.Modelo); err != nil {
			return nil, 0, err
		}
		itens = append(itens, item)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return itens, total, nil
}

// Consultar monta o detalhe consolidado da OS: cliente, veiculo, problemas, orcamentos com itens
// e a trilha de auditoria (`eventos`). Sem orcamento, sem problema ou sem evento nao e erro: a
// consulta devolve as listas vazias.
func (repository PostgresRepository) Consultar(ctx context.Context, ordemServicoID, clienteID string) (domain.ConsultaDetalhada, error) {
	consulta := domain.ConsultaDetalhada{OrdemServicoID: ordemServicoID, Problemas: []domain.ProblemaConsulta{}, Orcamentos: []domain.OrcamentoConsulta{}, Eventos: []domain.EventoConsulta{}}
	err := repository.db.QueryRow(ctx, `
		SELECT os.status, c.id, c.nome, c.documento, v.id, v.placa, v.marca, v.modelo, v.ano
		FROM ordem_servico os
		JOIN cliente c ON c.id = os.cliente_id
		JOIN veiculo v ON v.id = os.veiculo_id
		WHERE os.id = $1`, ordemServicoID,
	).Scan(&consulta.StatusOrdemServico, &consulta.Cliente.ID, &consulta.Cliente.Nome, &consulta.Cliente.Documento,
		&consulta.Veiculo.ID, &consulta.Veiculo.Placa, &consulta.Veiculo.Marca, &consulta.Veiculo.Modelo, &consulta.Veiculo.Ano)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConsultaDetalhada{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return domain.ConsultaDetalhada{}, err
	}
	if clienteID != "" && clienteID != consulta.Cliente.ID {
		return domain.ConsultaDetalhada{}, application.ErrAcessoNegado
	}

	problemaRows, err := repository.db.Query(ctx, `
		SELECT id, descricao, COALESCE(orcamento_id::text, ''), registrado_em
		FROM problema_ordem_servico WHERE ordem_servico_id = $1 ORDER BY registrado_em`, ordemServicoID)
	if err != nil {
		return domain.ConsultaDetalhada{}, err
	}
	for problemaRows.Next() {
		var problema domain.ProblemaConsulta
		if err = problemaRows.Scan(&problema.ID, &problema.Descricao, &problema.OrcamentoID, &problema.IdentificadoEm); err != nil {
			problemaRows.Close()
			return domain.ConsultaDetalhada{}, err
		}
		consulta.Problemas = append(consulta.Problemas, problema)
	}
	problemaRows.Close()
	if err = problemaRows.Err(); err != nil {
		return domain.ConsultaDetalhada{}, err
	}

	orcamentoRows, err := repository.db.Query(ctx, `
		SELECT id, tipo_orcamento, COALESCE(orcamento_original_id::text, ''), criado_em
		FROM orcamento WHERE ordem_servico_id = $1
		ORDER BY CASE tipo_orcamento WHEN 'PRINCIPAL' THEN 0 ELSE 1 END, criado_em`, ordemServicoID)
	if err != nil {
		return domain.ConsultaDetalhada{}, err
	}
	for orcamentoRows.Next() {
		var orcamento domain.OrcamentoConsulta
		if err = orcamentoRows.Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.OrcamentoOriginalID, &orcamento.DataGeracao); err != nil {
			orcamentoRows.Close()
			return domain.ConsultaDetalhada{}, err
		}
		consulta.Orcamentos = append(consulta.Orcamentos, orcamento)
	}
	orcamentoRows.Close()
	if err = orcamentoRows.Err(); err != nil {
		return domain.ConsultaDetalhada{}, err
	}
	for index := range consulta.Orcamentos {
		itens, valorTotal, err := itensDoOrcamentoOS(ctx, repository.db, consulta.Orcamentos[index].ID)
		if err != nil {
			return domain.ConsultaDetalhada{}, err
		}
		consulta.Orcamentos[index].Itens = itens
		consulta.Orcamentos[index].ValorTotal = valorTotal
		consulta.ValorTotalGeral += valorTotal
	}

	eventoRows, err := repository.db.Query(ctx, `
		SELECT id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em, registrado_em
		FROM auditoria_ordem_servico WHERE ordem_servico_id = $1 ORDER BY ocorrido_em`, ordemServicoID)
	if err != nil {
		return domain.ConsultaDetalhada{}, err
	}
	for eventoRows.Next() {
		var evento domain.EventoConsulta
		if err = eventoRows.Scan(&evento.ID, &evento.Agregado, &evento.AgregadoID, &evento.TipoEvento,
			&evento.Dados, &evento.Metadados, &evento.OcorridoEm, &evento.RegistradoEm); err != nil {
			eventoRows.Close()
			return domain.ConsultaDetalhada{}, err
		}
		consulta.Eventos = append(consulta.Eventos, evento)
	}
	eventoRows.Close()
	if err = eventoRows.Err(); err != nil {
		return domain.ConsultaDetalhada{}, err
	}

	return consulta, nil
}

func itensDoOrcamentoOS(ctx context.Context, db *pgxpool.Pool, orcamentoID string) ([]domain.ItemOrcamentoConsulta, float64, error) {
	rows, err := db.Query(ctx, `
		SELECT tipo_item, descricao, quantidade, valor_unitario, valor_total
		FROM orcamento_item WHERE orcamento_id = $1 ORDER BY id`, orcamentoID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	itens := []domain.ItemOrcamentoConsulta{}
	var total float64
	for rows.Next() {
		var item domain.ItemOrcamentoConsulta
		if err = rows.Scan(&item.Tipo, &item.Descricao, &item.Quantidade, &item.ValorUnitario, &item.ValorTotal); err != nil {
			return nil, 0, err
		}
		total += item.ValorTotal
		itens = append(itens, item)
	}
	return itens, total, rows.Err()
}

// Finalizar bloqueia enquanto houver servico pendente, orcamento complementar em aberto ou
// reserva ativa sem baixa. A notificacao ao cliente e best-effort, apos o commit: nao ha
// provedor de e-mail configurado neste repositorio, entao o envio e apenas registrado em log.
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

func (repository PostgresRepository) RegistrarProblemaRelatado(ctx context.Context, ordemServicoID string, problema domain.ProblemaRelatado) (resultado domain.OrdemDeServico, err error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return resultado, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	var descricaoExistente *string
	err = tx.QueryRow(ctx, `SELECT status, problema_relatado_descricao
		FROM ordem_servico WHERE id = $1 FOR UPDATE`, ordemServicoID).Scan(&status, &descricaoExistente)
	if errors.Is(err, pgx.ErrNoRows) {
		return resultado, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return resultado, err
	}
	if descricaoExistente != nil {
		return resultado, application.ErrProblemaRelatadoJaRegistrado
	}
	if status != domain.StatusRecebida {
		return resultado, application.ErrOrdemServicoForaDeRecebida
	}

	var inicio time.Time
	err = tx.QueryRow(ctx, `UPDATE ordem_servico
		SET problema_relatado_descricao = $2,
			problema_relatado_observacoes = NULLIF($3, ''),
			data_inicio_diagnostico = CURRENT_TIMESTAMP,
			status = $4
		WHERE id = $1
		RETURNING id, status, data_inicio_diagnostico`, ordemServicoID, problema.Descricao, problema.Observacoes, domain.StatusEmDiagnostico).
		Scan(&resultado.ID, &resultado.Status, &inicio)
	if err != nil {
		return resultado, err
	}
	resultado.ProblemaRelatado = problema
	resultado.DataInicioDiagnostico = &inicio
	if err = tx.Commit(ctx); err != nil {
		return domain.OrdemDeServico{}, err
	}
	return resultado, nil
}

func (repository PostgresRepository) Criar(ctx context.Context, input application.CriarInput) (ordem domain.OrdemDeServico, err error) {
	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ordem, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var clienteID string
	err = tx.QueryRow(ctx, `SELECT id FROM cliente WHERE id = $1 AND ativo FOR SHARE`, input.ClienteID).Scan(&clienteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordem, application.ErrClienteNaoEncontrado
	}
	if err != nil {
		return ordem, err
	}

	var veiculoClienteID, placa string
	err = tx.QueryRow(ctx, `SELECT cliente_id, placa FROM veiculo WHERE id = $1 AND ativo FOR SHARE`, input.VeiculoID).Scan(&veiculoClienteID, &placa)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordem, application.ErrVeiculoNaoEncontrado
	}
	if err != nil {
		return ordem, err
	}
	if veiculoClienteID != input.ClienteID {
		return ordem, application.ErrVeiculoNaoVinculadoCliente
	}

	err = tx.QueryRow(ctx, `INSERT INTO ordem_servico (cliente_id, veiculo_id, placa_veiculo, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, cliente_id, veiculo_id, placa_veiculo, status, criada_em`, input.ClienteID, input.VeiculoID, placa, domain.StatusRecebida).
		Scan(&ordem.ID, &ordem.ClienteID, &ordem.VeiculoID, &ordem.PlacaVeiculo, &ordem.Status, &ordem.CriadaEm)
	if err != nil {
		return ordem, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.OrdemDeServico{}, err
	}
	return ordem, nil
}

func (repository PostgresRepository) RegistrarItens(ctx context.Context, input application.RegistrarInput) (domainOrcamento.Resultado, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domainOrcamento.Resultado{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", input.OSID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainOrcamento.Resultado{}, application.ErrOSNaoEncontrada
	}
	if err != nil {
		return domainOrcamento.Resultado{}, err
	}
	if !domain.PermiteRegistroDeItens(status) {
		return domainOrcamento.Resultado{}, domain.ErrStatusNaoPermiteItens
	}

	budgetID, originalID, budgetType, budgetStatus, err := repository.resolveBudget(ctx, tx, input.OSID, status)
	if err != nil {
		return domainOrcamento.Resultado{}, err
	}

	result := domainOrcamento.Resultado{OrdemServicoID: input.OSID, OrcamentoID: budgetID, OrcamentoOriginal: originalID, TipoOrcamento: budgetType, StatusOrcamento: budgetStatus, RegistradoPor: input.UsuarioID}
	for _, requested := range input.Itens {
		item, err := repository.loadItem(ctx, tx, requested.ItemID)
		if err != nil {
			return domainOrcamento.Resultado{}, err
		}
		if !item.ativo {
			return domainOrcamento.Resultado{}, application.ErrItemInativo
		}
		if item.tipo != input.Tipo {
			return domainOrcamento.Resultado{}, domainEstoque.ErrTipoItemInvalido
		}
		if err := domainEstoque.QuantidadeCompativelComUnidade(requested.Quantidade, item.unidadeMedida); err != nil {
			return domainOrcamento.Resultado{}, err
		}
		if item.valor == nil {
			return domainOrcamento.Resultado{}, application.ErrItemSemValor
		}

		itemValue := *item.valor
		if _, err = tx.Exec(ctx, `
			INSERT INTO ordem_servico_item (ordem_servico_id, item_estoque_id, quantidade_necessaria, valor_unitario)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (ordem_servico_id, item_estoque_id)
			DO UPDATE SET quantidade_necessaria = ordem_servico_item.quantidade_necessaria + EXCLUDED.quantidade_necessaria,
			              valor_unitario = EXCLUDED.valor_unitario`, input.OSID, requested.ItemID, requested.Quantidade, itemValue); err != nil {
			return domainOrcamento.Resultado{}, fmt.Errorf("inserir item da os: %w", err)
		}
		valueTotal := requested.Quantidade * itemValue
		if _, err = tx.Exec(ctx, `
			INSERT INTO orcamento_item (orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`, budgetID, requested.ItemID, input.Tipo, item.descricao, requested.Quantidade, itemValue, valueTotal); err != nil {
			return domainOrcamento.Resultado{}, fmt.Errorf("inserir item do orcamento: %w", err)
		}
		result.ItensRegistrados = append(result.ItensRegistrados, domainOrcamento.ItemRegistrado{ItemID: requested.ItemID, Codigo: item.codigo, Descricao: item.descricao, Tipo: item.tipo, Quantidade: requested.Quantidade, ValorUnitario: itemValue, ValorItem: valueTotal})
	}

	if err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(valor_total), 0) FROM orcamento_item WHERE orcamento_id = $1", budgetID).Scan(&result.ValorOrcamento); err != nil {
		return domainOrcamento.Resultado{}, fmt.Errorf("calcular valor do orcamento: %w", err)
	}
	if err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(oi.valor_total), 0) FROM orcamento_item oi JOIN orcamento o ON o.id = oi.orcamento_id WHERE o.ordem_servico_id = $1`, input.OSID).Scan(&result.ValorTotalGeral); err != nil {
		return domainOrcamento.Resultado{}, fmt.Errorf("calcular total da os: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em) VALUES ($1, NULLIF($2, '')::uuid, 'ORCAMENTO', $3, 'ITENS_REGISTRADOS', jsonb_build_object('tipoItem', $4::text, 'quantidadeItens', $5::integer), '{}'::jsonb, CURRENT_TIMESTAMP)`, input.OSID, input.UsuarioID, budgetID, input.Tipo, len(input.Itens)); err != nil {
		return domainOrcamento.Resultado{}, fmt.Errorf("registrar auditoria: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return domainOrcamento.Resultado{}, err
	}
	return result, nil
}

func (repository PostgresRepository) resolveBudget(ctx context.Context, tx pgx.Tx, osID, status string) (string, string, string, string, error) {
	var id, original, budgetType, budgetStatus string
	err := tx.QueryRow(ctx, "SELECT id, COALESCE(orcamento_original_id::text, ''), tipo_orcamento, status FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = 'PRINCIPAL'", osID).Scan(&id, &original, &budgetType, &budgetStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		if domain.EhComplementar(status) {
			return "", "", "", "", application.ErrOrcamentoNaoEncontrado
		}
		err = tx.QueryRow(ctx, "INSERT INTO orcamento (ordem_servico_id, tipo_orcamento, status) VALUES ($1, 'PRINCIPAL', 'CRIADO') RETURNING id", osID).Scan(&id)
		return id, "", domainOrcamento.TipoPrincipal, domainOrcamento.StatusCriado, err
	}
	if err != nil {
		return "", "", "", "", err
	}
	if domain.EhComplementar(status) {
		if budgetStatus != "APROVADO" {
			return "", "", "", "", application.ErrOrcamentoAprovado
		}
		principalID := id
		err = tx.QueryRow(ctx, "INSERT INTO orcamento (ordem_servico_id, orcamento_original_id, tipo_orcamento, status) VALUES ($1, $2, 'COMPLEMENTAR', 'CRIADO') RETURNING id", osID, principalID).Scan(&id)
		return id, principalID, domainOrcamento.TipoComplementar, domainOrcamento.StatusCriado, err
	}
	if budgetStatus == "APROVADO" {
		return "", "", "", "", application.ErrOrcamentoAprovado
	}
	return id, original, budgetType, budgetStatus, nil
}

type itemRow struct {
	tipo, codigo, descricao, unidadeMedida string
	ativo                                  bool
	valor                                  *float64
}

func (repository PostgresRepository) loadItem(ctx context.Context, tx pgx.Tx, itemID string) (itemRow, error) {
	var item itemRow
	var preco, custo sql.NullFloat64
	err := tx.QueryRow(ctx, "SELECT tipo, codigo, descricao, unidade_medida, ativo, preco_venda, custo_unitario FROM item_estoque WHERE id = $1", itemID).Scan(&item.tipo, &item.codigo, &item.descricao, &item.unidadeMedida, &item.ativo, &preco, &custo)
	if errors.Is(err, pgx.ErrNoRows) {
		return itemRow{}, application.ErrItemNaoEncontrado
	}
	if err != nil {
		return itemRow{}, err
	}
	if item.tipo == domainEstoque.TipoPeca && preco.Valid {
		item.valor = &preco.Float64
	}
	if item.tipo == domainEstoque.TipoInsumo && custo.Valid {
		item.valor = &custo.Float64
	}
	return item, nil
}

func (repository PostgresRepository) RegistrarProblema(ctx context.Context, ordemServicoID string, cadastro domain.ProblemaCadastro) (application.ResultadoRegistroProblema, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", ordemServicoID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ResultadoRegistroProblema{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	tipo, err := domain.TipoOrcamentoParaStatus(status)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}

	orcamento, err := obterOuCriarOrcamento(ctx, tx, ordemServicoID, tipo)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	problema := domain.Problema{Descricao: cadastro.Descricao, Observacoes: cadastro.Observacoes}
	err = tx.QueryRow(ctx, `
		INSERT INTO problema_ordem_servico (ordem_servico_id, orcamento_id, descricao, observacoes)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id, registrado_em`, ordemServicoID, orcamento.ID, cadastro.Descricao, cadastro.Observacoes,
	).Scan(&problema.ID, &problema.RegistradoEm)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	return application.ResultadoRegistroProblema{Problema: problema, Orcamento: orcamento}, nil
}

func (repository PostgresRepository) RegistrarServicos(ctx context.Context, ordemServicoID string, cadastros []domain.ServicoCadastro) (application.ResultadoRegistroServicos, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", ordemServicoID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ResultadoRegistroServicos{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	tipo, err := domain.TipoOrcamentoParaServico(status)
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}

	var orcamento domain.Orcamento
	err = tx.QueryRow(ctx, `
		SELECT id, tipo_orcamento, status
		FROM orcamento
		WHERE ordem_servico_id = $1 AND tipo_orcamento = $2 AND status = 'CRIADO'
		ORDER BY criado_em DESC LIMIT 1 FOR UPDATE`, ordemServicoID, tipo,
	).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ResultadoRegistroServicos{}, domain.ErrOrcamentoAplicavelNaoEncontrado
	}
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}

	resultado := application.ResultadoRegistroServicos{Orcamento: orcamento, Servicos: make([]domain.ServicoRegistrado, 0, len(cadastros))}
	for _, cadastro := range cadastros {
		var descricao string
		var valor float64
		var ativo bool
		err = tx.QueryRow(ctx, "SELECT nome, valor, ativo FROM servico WHERE id = $1", cadastro.ServicoID).Scan(&descricao, &valor, &ativo)
		if errors.Is(err, pgx.ErrNoRows) {
			return application.ResultadoRegistroServicos{}, domain.ErrServicoNaoEncontrado
		}
		if err != nil {
			return application.ResultadoRegistroServicos{}, err
		}
		if !ativo {
			return application.ResultadoRegistroServicos{}, domain.ErrServicoInativo
		}
		var existe bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM orcamento_item WHERE orcamento_id = $1 AND servico_id = $2)", orcamento.ID, cadastro.ServicoID).Scan(&existe); err != nil {
			return application.ResultadoRegistroServicos{}, err
		}
		if existe {
			return application.ResultadoRegistroServicos{}, domain.ErrServicoDuplicado
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO orcamento_item (orcamento_id, servico_id, tipo_item, descricao, quantidade, valor_unitario, valor_total, observacao)
			VALUES ($1, $2, 'SERVICO', $3, 1, $4, $4, NULLIF($5, ''))`, orcamento.ID, cadastro.ServicoID, descricao, valor, cadastro.Observacao,
		); err != nil {
			return application.ResultadoRegistroServicos{}, err
		}
		resultado.Servicos = append(resultado.Servicos, domain.ServicoRegistrado{ServicoID: cadastro.ServicoID, Descricao: descricao, ValorUnitario: valor, Observacao: cadastro.Observacao})
	}
	if err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(valor_total), 0) FROM orcamento_item WHERE orcamento_id = $1", orcamento.ID).Scan(&resultado.Orcamento.ValorTotal); err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	if _, err = tx.Exec(ctx, "UPDATE orcamento SET data_atualizacao = CURRENT_TIMESTAMP WHERE id = $1", orcamento.ID); err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	return resultado, nil
}

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

func obterOuCriarOrcamento(ctx context.Context, tx pgx.Tx, ordemServicoID, tipo string) (domain.Orcamento, error) {
	query := "SELECT id, tipo_orcamento, status FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = $2"
	args := []any{ordemServicoID, tipo}
	if tipo == domain.OrcamentoComplementar {
		query += " AND status = 'CRIADO' ORDER BY criado_em DESC LIMIT 1"
	} else {
		query += " LIMIT 1"
	}
	var orcamento domain.Orcamento
	err := tx.QueryRow(ctx, query, args...).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	if err == nil {
		if orcamento.Status != domain.OrcamentoCriado {
			return domain.Orcamento{}, domain.ErrOrcamentoFechado
		}
		return orcamento, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Orcamento{}, err
	}
	if tipo == domain.OrcamentoComplementar {
		var principalID string
		err = tx.QueryRow(ctx, "SELECT id FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = 'PRINCIPAL'", ordemServicoID).Scan(&principalID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Orcamento{}, domain.ErrOrcamentoPrincipalNaoEncontrado
		}
		if err != nil {
			return domain.Orcamento{}, err
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO orcamento (ordem_servico_id, orcamento_original_id, tipo_orcamento, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id, tipo_orcamento, status`, ordemServicoID, principalID, tipo, domain.OrcamentoCriado,
		).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
		return orcamento, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO orcamento (ordem_servico_id, tipo_orcamento, status)
		VALUES ($1, $2, $3)
		RETURNING id, tipo_orcamento, status`, ordemServicoID, tipo, domain.OrcamentoCriado,
	).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	return orcamento, err
}
