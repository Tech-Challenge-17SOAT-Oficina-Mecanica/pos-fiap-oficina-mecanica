package ordemservico

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

func (repository PostgresRepository) ConsultarTempoExecucao(ctx context.Context, ordemServicoID string) (domain.TempoExecucao, error) {
	var dataInicio, dataFinalizacao *time.Time
	err := repository.db.QueryRow(ctx, `
		SELECT iniciada_em, finalizada_em
		FROM ordem_servico
		WHERE id = $1`, ordemServicoID,
	).Scan(&dataInicio, &dataFinalizacao)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TempoExecucao{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return domain.TempoExecucao{}, err
	}
	if dataInicio == nil || dataFinalizacao == nil {
		return domain.TempoExecucao{}, domain.ErrTempoExecucaoIndisponivel
	}
	return domain.NovoTempoExecucao(ordemServicoID, *dataInicio, *dataFinalizacao)
}

func (repository PostgresRepository) ListarTemposExecucao(ctx context.Context, dataInicio, dataFim *time.Time, limite, deslocamento int) ([]domain.TempoExecucao, int, int64, error) {
	const filtro = `
		FROM ordem_servico
		WHERE iniciada_em IS NOT NULL
		  AND finalizada_em IS NOT NULL
		  AND finalizada_em >= iniciada_em
		  AND ($1::timestamptz IS NULL OR iniciada_em >= $1)
		  AND ($2::timestamptz IS NULL OR iniciada_em < $2 + INTERVAL '1 day')`
	var total int
	var totalMinutos int64
	if err := repository.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(EXTRACT(EPOCH FROM finalizada_em - iniciada_em) / 60)::bigint, 0)`+filtro,
		dataInicio, dataFim,
	).Scan(&total, &totalMinutos); err != nil {
		return nil, 0, 0, err
	}
	rows, err := repository.db.Query(ctx, `
		SELECT id, iniciada_em, finalizada_em`+filtro+`
		ORDER BY iniciada_em DESC, id DESC
		LIMIT $3 OFFSET $4`, dataInicio, dataFim, limite, deslocamento)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	itens := make([]domain.TempoExecucao, 0)
	for rows.Next() {
		var ordemServicoID string
		var inicio, fim time.Time
		if err = rows.Scan(&ordemServicoID, &inicio, &fim); err != nil {
			return nil, 0, 0, err
		}
		tempoExecucao, err := domain.NovoTempoExecucao(ordemServicoID, inicio, fim)
		if err != nil {
			return nil, 0, 0, err
		}
		itens = append(itens, tempoExecucao)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, 0, err
	}
	return itens, total, totalMinutos, nil
}

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

	// $4 traz os status listaveis em ordem de prioridade, definidos no dominio: o mesmo array
	// exclui as OS encerradas e ordena as demais, mais antigas primeiro dentro de cada status.
	const filtro = `
		FROM ordem_servico os
		JOIN cliente c ON c.id = os.cliente_id
		JOIN veiculo v ON v.id = os.veiculo_id
		WHERE ($1 = '' OR os.status = $1)
		  AND ($2 = '' OR c.documento = $2)
		  AND ($3 = '' OR v.placa = $3)
		  AND os.status = ANY($4::text[])`

	prioridade := domain.StatusListagemPorPrioridade
	var total int
	if err := repository.db.QueryRow(ctx, "SELECT COUNT(*) "+filtro, filtros.Status, filtros.Documento, filtros.Placa, prioridade).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := repository.db.Query(ctx, `
		SELECT os.id, os.status, os.criada_em, c.id, c.nome, c.documento, v.id, v.placa, v.marca, v.modelo`+filtro+`
		ORDER BY array_position($4::text[], os.status::text), os.criada_em ASC, os.id ASC
		LIMIT $5 OFFSET $6`, filtros.Status, filtros.Documento, filtros.Placa, prioridade, limite, deslocamento)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	itens := []domain.ItemListagem{}
	for rows.Next() {
		var item domain.ItemListagem
		if err = rows.Scan(&item.OrdemServicoID, &item.Status, &item.CriadaEm, &item.ClienteID, &item.ClienteNome, &item.ClienteDocumento,
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
