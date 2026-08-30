package estoque

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domainOS "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

const operacaoRegistrarSaida = "SAIDA_ESTOQUE"

type itemSaidaRow struct {
	id, codigo, tipo, unidadeMedida string
	saldoFisico, saldoReservado     float64
	custoUnitario                   float64
}

type reservaSaidaRow struct {
	id, osItemID string
	quantidade   float64
}

func (repository PostgresRepository) RegistrarSaida(ctx context.Context, input application.RegistrarSaidaInput, cadastro domain.SaidaCadastro) (application.ResultadoSaida, error) {
	hash := hashRequisicaoSaida(input)
	if resultado, encontrada, err := carregarRespostaSaida(ctx, repository.db, input.IdempotencyKey); err != nil {
		return application.ResultadoSaida{}, err
	} else if encontrada {
		return application.ResultadoSaida{Saida: resultado, JaProcessada: true}, nil
	}

	resultado, err := repository.processarSaida(ctx, input, cadastro)
	if err != nil {
		return application.ResultadoSaida{}, err
	}
	if err = gravarChaveSaida(ctx, repository.db, input.IdempotencyKey, hash, resultado); err != nil {
		if isUniqueViolation(err) {
			existente, encontrada, ferr := carregarRespostaSaida(ctx, repository.db, input.IdempotencyKey)
			if ferr != nil {
				return application.ResultadoSaida{}, ferr
			}
			if encontrada {
				return application.ResultadoSaida{Saida: existente, JaProcessada: true}, nil
			}
		}
		return application.ResultadoSaida{}, err
	}
	return application.ResultadoSaida{Saida: resultado}, nil
}

func (repository PostgresRepository) processarSaida(ctx context.Context, input application.RegistrarSaidaInput, cadastro domain.SaidaCadastro) (domain.ResultadoSaida, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.ResultadoSaida{}, err
	}
	defer tx.Rollback(ctx)

	var statusOS string
	if err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", cadastro.OrdemServicoID).Scan(&statusOS); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ResultadoSaida{}, application.ErrOrdemServicoNaoEncontrada
		}
		return domain.ResultadoSaida{}, err
	}
	if statusOS != domainOS.StatusEmExecucao {
		return domain.ResultadoSaida{}, application.ErrOSForaDeExecucao
	}

	itens, err := carregarItensSaida(ctx, tx, cadastro.Itens)
	if err != nil {
		return domain.ResultadoSaida{}, err
	}

	saidaID, registradoEm, err := novaSaidaID(ctx, tx)
	if err != nil {
		return domain.ResultadoSaida{}, err
	}
	resultado := domain.ResultadoSaida{
		SaidaID: saidaID, OrdemServicoID: cadastro.OrdemServicoID, RegistradoEm: registradoEm, RegistradoPor: input.UsuarioID,
	}

	for _, requisitado := range cadastro.Itens {
		item := itens[requisitado.ItemID]
		reservas, quantidadeReservada, osItemID, err := carregarReservasAtivas(ctx, tx, cadastro.OrdemServicoID, item.id)
		if err != nil {
			return domain.ResultadoSaida{}, err
		}
		if len(reservas) == 0 {
			return domain.ResultadoSaida{}, application.ErrReservaAtivaNaoEncontrada
		}
		if requisitado.Quantidade > quantidadeReservada {
			return domain.ResultadoSaida{}, domain.ErrQuantidadeMaiorQueReserva
		}
		quantidadeLiberada := 0.0
		if cadastro.LiberarSaldoNaoUsado {
			quantidadeLiberada = quantidadeReservada - requisitado.Quantidade
		}
		baixaReserva := requisitado.Quantidade + quantidadeLiberada
		if item.saldoFisico < requisitado.Quantidade || item.saldoReservado < baixaReserva {
			return domain.ResultadoSaida{}, application.ErrSaldoInsuficiente
		}

		var saldoFisicoAtual, saldoReservadoAtual float64
		err = tx.QueryRow(ctx, `
			UPDATE item_estoque
			SET saldo_fisico = saldo_fisico - $2::numeric,
			    saldo_reservado = saldo_reservado - $3::numeric
			WHERE id = $1
			  AND saldo_fisico >= $2::numeric
			  AND saldo_reservado >= $3::numeric
			RETURNING saldo_fisico, saldo_reservado`, item.id, requisitado.Quantidade, baixaReserva,
		).Scan(&saldoFisicoAtual, &saldoReservadoAtual)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ResultadoSaida{}, application.ErrSaldoInsuficiente
		}
		if err != nil {
			return domain.ResultadoSaida{}, err
		}
		if _, err = tx.Exec(ctx, `
			UPDATE ordem_servico_item
			SET quantidade_reservada = quantidade_reservada - $2::numeric,
			    quantidade_consumida = quantidade_consumida + $3::numeric
			WHERE id = $1`, osItemID, baixaReserva, requisitado.Quantidade,
		); err != nil {
			return domain.ResultadoSaida{}, err
		}
		if err = consumirReservas(ctx, tx, reservas, requisitado.Quantidade, cadastro.LiberarSaldoNaoUsado); err != nil {
			return domain.ResultadoSaida{}, err
		}

		custoTotal := requisitado.Quantidade * item.custoUnitario
		if _, err = tx.Exec(ctx, `
			INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, reserva_estoque_id, tipo, quantidade, custo_unitario, documento_origem)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			item.id, cadastro.OrdemServicoID, reservas[0].id, domain.MovimentacaoSaida, requisitado.Quantidade, item.custoUnitario, "SAIDA-"+saidaID,
		); err != nil {
			return domain.ResultadoSaida{}, err
		}
		if quantidadeLiberada > 0 {
			if _, err = tx.Exec(ctx, `
				INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, reserva_estoque_id, tipo, quantidade, documento_origem)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				item.id, cadastro.OrdemServicoID, reservas[0].id, domain.MovimentacaoLiberacaoReserva, quantidadeLiberada, "LIBERACAO-"+saidaID,
			); err != nil {
				return domain.ResultadoSaida{}, err
			}
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO auditoria_estoque (item_estoque_id, usuario_id, tipo_evento, documento_origem, dados, ocorrido_em)
			VALUES ($1, NULLIF($2, '')::uuid, 'SAIDA_ESTOQUE', $3,
			        jsonb_build_object('ordemServicoId', $4::text, 'quantidadeBaixada', $5::numeric, 'quantidadeLiberada', $6::numeric),
			        $7)`,
			item.id, input.UsuarioID, "SAIDA-"+saidaID, cadastro.OrdemServicoID, requisitado.Quantidade, quantidadeLiberada, registradoEm,
		); err != nil {
			return domain.ResultadoSaida{}, err
		}

		resultado.CustoTotalSaida += custoTotal
		resultado.Itens = append(resultado.Itens, domain.ItemSaidaResultado{
			ItemID: item.id, Codigo: item.codigo, Tipo: item.tipo, UnidadeMedida: item.unidadeMedida,
			QuantidadeBaixada: requisitado.Quantidade, QuantidadeReservadaAntes: quantidadeReservada,
			QuantidadeLiberada: quantidadeLiberada, SaldoFisicoAtual: saldoFisicoAtual, SaldoReservadoAtual: saldoReservadoAtual,
			CustoUnitario: item.custoUnitario, CustoTotal: custoTotal,
		})
	}

	if _, err = tx.Exec(ctx, "UPDATE ordem_servico SET custo_total_materiais = custo_total_materiais + $2::numeric WHERE id = $1", cadastro.OrdemServicoID, resultado.CustoTotalSaida); err != nil {
		return domain.ResultadoSaida{}, err
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em)
		VALUES ($1, NULLIF($2, '')::uuid, 'ORDEM_SERVICO', $1, 'SAIDA_ESTOQUE',
		        jsonb_build_object('saidaId', $3::text, 'custoTotalSaida', $4::numeric),
		        '{}'::jsonb, $5)`,
		cadastro.OrdemServicoID, input.UsuarioID, saidaID, resultado.CustoTotalSaida, registradoEm,
	); err != nil {
		return domain.ResultadoSaida{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ResultadoSaida{}, err
	}
	return resultado, nil
}

func carregarItensSaida(ctx context.Context, tx pgx.Tx, itens []domain.ItemSaida) (map[string]itemSaidaRow, error) {
	ids := make([]string, len(itens))
	for index, item := range itens {
		ids[index] = item.ItemID
	}
	rows, err := tx.Query(ctx, `
		SELECT id, codigo, tipo, unidade_medida, saldo_fisico, saldo_reservado, COALESCE(custo_unitario, 0)
		FROM item_estoque WHERE id = ANY($1) ORDER BY id FOR UPDATE`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	carregados := make(map[string]itemSaidaRow, len(itens))
	for rows.Next() {
		var item itemSaidaRow
		if err = rows.Scan(&item.id, &item.codigo, &item.tipo, &item.unidadeMedida, &item.saldoFisico, &item.saldoReservado, &item.custoUnitario); err != nil {
			return nil, err
		}
		carregados[item.id] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for _, solicitado := range itens {
		item, ok := carregados[solicitado.ItemID]
		if !ok {
			return nil, application.ErrItemNaoEncontrado
		}
		if err = domain.QuantidadeValida(item.tipo, solicitado.Quantidade); err != nil {
			return nil, err
		}
		if item.tipo == domain.TipoInsumo {
			if err = domain.QuantidadeCompativelComUnidade(solicitado.Quantidade, item.unidadeMedida); err != nil {
				return nil, err
			}
		}
	}
	return carregados, nil
}

func carregarReservasAtivas(ctx context.Context, tx pgx.Tx, ordemServicoID, itemID string) ([]reservaSaidaRow, float64, string, error) {
	rows, err := tx.Query(ctx, `
		SELECT r.id, r.ordem_servico_item_id, r.quantidade
		FROM reserva_estoque r
		JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id
		WHERE osi.ordem_servico_id = $1 AND r.item_estoque_id = $2 AND r.status = $3
		ORDER BY r.id FOR UPDATE OF r`, ordemServicoID, itemID, domain.ReservaAtiva)
	if err != nil {
		return nil, 0, "", err
	}
	defer rows.Close()
	var reservas []reservaSaidaRow
	var total float64
	var osItemID string
	for rows.Next() {
		var reserva reservaSaidaRow
		if err = rows.Scan(&reserva.id, &reserva.osItemID, &reserva.quantidade); err != nil {
			return nil, 0, "", err
		}
		total += reserva.quantidade
		osItemID = reserva.osItemID
		reservas = append(reservas, reserva)
	}
	return reservas, total, osItemID, rows.Err()
}

func consumirReservas(ctx context.Context, tx pgx.Tx, reservas []reservaSaidaRow, quantidade float64, liberarSaldoNaoUsado bool) error {
	if liberarSaldoNaoUsado {
		for _, reserva := range reservas {
			if _, err := tx.Exec(ctx, "UPDATE reserva_estoque SET status = $2, liberada_em = CURRENT_TIMESTAMP WHERE id = $1", reserva.id, domain.ReservaConsumida); err != nil {
				return err
			}
		}
		return nil
	}
	restante := quantidade
	for _, reserva := range reservas {
		if restante <= 0 {
			break
		}
		if restante >= reserva.quantidade {
			if _, err := tx.Exec(ctx, "UPDATE reserva_estoque SET status = $2, liberada_em = CURRENT_TIMESTAMP WHERE id = $1", reserva.id, domain.ReservaConsumida); err != nil {
				return err
			}
			restante -= reserva.quantidade
			continue
		}
		if _, err := tx.Exec(ctx, "UPDATE reserva_estoque SET quantidade = quantidade - $2::numeric WHERE id = $1", reserva.id, restante); err != nil {
			return err
		}
		restante = 0
	}
	return nil
}

func novaSaidaID(ctx context.Context, tx pgx.Tx) (string, time.Time, error) {
	var id string
	var agora time.Time
	err := tx.QueryRow(ctx, "SELECT gen_random_uuid()::text, CURRENT_TIMESTAMP").Scan(&id, &agora)
	return id, agora, err
}

func hashRequisicaoSaida(input application.RegistrarSaidaInput) string {
	soma := sha256.Sum256([]byte(fmt.Sprintf("%s|%t|%v|%s", input.OrdemServicoID, input.LiberarSaldoNaoUsado, input.Itens, input.UsuarioID)))
	return hex.EncodeToString(soma[:])
}

func carregarRespostaSaida(ctx context.Context, db *pgxpool.Pool, chave string) (domain.ResultadoSaida, bool, error) {
	var resposta []byte
	err := db.QueryRow(ctx, "SELECT resposta FROM chave_idempotencia WHERE operacao = $1 AND chave = $2", operacaoRegistrarSaida, chave).Scan(&resposta)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ResultadoSaida{}, false, nil
	}
	if err != nil {
		return domain.ResultadoSaida{}, false, err
	}
	var resultado domain.ResultadoSaida
	if err = json.Unmarshal(resposta, &resultado); err != nil {
		return domain.ResultadoSaida{}, false, err
	}
	return resultado, true, nil
}

func gravarChaveSaida(ctx context.Context, db *pgxpool.Pool, chave, hash string, resultado domain.ResultadoSaida) error {
	payload, err := json.Marshal(resultado)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
		INSERT INTO chave_idempotencia (chave, operacao, hash_requisicao, status_resposta, resposta)
		VALUES ($1, $2, $3, 201, $4)`, chave, operacaoRegistrarSaida, hash, payload)
	return err
}
