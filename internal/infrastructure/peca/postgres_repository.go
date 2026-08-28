package peca

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

const colunas = `i.id, i.codigo, i.nome, i.descricao, i.categoria_id, c.nome,
	i.fabricante, i.unidade_medida, i.preco_venda::text,
	i.saldo_fisico, i.saldo_reservado, i.estoque_minimo, i.ativo, i.version,
	EXISTS (
		SELECT 1 FROM pedido_compra_item pci
		JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
		WHERE pci.item_estoque_id = i.id AND pc.status IN ('ABERTO', 'PARCIAL')
	)`

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

func (repository PostgresRepository) BuscarPorFiltro(ctx context.Context, filtros pecaApplication.Filtros, limite, deslocamento int) ([]peca.Peca, int, error) {
	condicoes, args := montarCondicoes(filtros)

	var total int
	contagem := "SELECT COUNT(*) FROM item_estoque i WHERE " + condicoes
	if err := repository.db.QueryRow(ctx, contagem, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	listagem := fmt.Sprintf(`SELECT %s FROM item_estoque i
		JOIN categoria c ON c.id = i.categoria_id
		WHERE %s ORDER BY i.codigo LIMIT $%d OFFSET $%d`,
		colunas, condicoes, len(args)+1, len(args)+2)

	rows, err := repository.db.Query(ctx, listagem, append(args, limite, deslocamento)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var pecas []peca.Peca
	for rows.Next() {
		item, err := ler(rows)
		if err != nil {
			return nil, 0, err
		}
		pecas = append(pecas, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return pecas, total, nil
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (peca.Peca, error) {
	consulta := fmt.Sprintf(`SELECT %s FROM item_estoque i
		JOIN categoria c ON c.id = i.categoria_id
		WHERE i.id = $1 AND i.tipo = 'PECA'`, colunas)

	item, err := ler(repository.db.QueryRow(ctx, consulta, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrNaoEncontrada
	}
	return item, err
}

type scanner interface {
	Scan(destinos ...any) error
}

func ler(linha scanner) (peca.Peca, error) {
	var item peca.Peca
	err := linha.Scan(
		&item.ID, &item.Codigo, &item.Nome, &item.Descricao, &item.CategoriaID, &item.Categoria,
		&item.Fabricante, &item.UnidadeMedida, &item.PrecoVenda,
		&item.SaldoFisico, &item.SaldoReservado, &item.EstoqueMinimo, &item.Ativo, &item.Version,
		&item.PossuiPedidoEmAberto,
	)
	return item, err
}

func montarCondicoes(filtros pecaApplication.Filtros) (string, []any) {
	condicoes := []string{"i.tipo = 'PECA'"}
	var args []any

	adicionar := func(formato string, valor any) {
		args = append(args, valor)
		condicoes = append(condicoes, fmt.Sprintf(formato, len(args)))
	}

	if !filtros.IncluirInativos {
		condicoes = append(condicoes, "i.ativo")
	}
	if filtros.Codigo != "" {
		adicionar("i.codigo = $%d", filtros.Codigo)
	}
	if filtros.Descricao != "" {
		adicionar("i.descricao ILIKE $%d", "%"+filtros.Descricao+"%")
	}
	if filtros.CategoriaID != "" {
		adicionar("i.categoria_id = $%d", filtros.CategoriaID)
	}
	if filtros.Fabricante != "" {
		adicionar("i.fabricante ILIKE $%d", "%"+filtros.Fabricante+"%")
	}
	if filtros.SomenteDisponiveis {
		condicoes = append(condicoes, "(i.saldo_fisico - i.saldo_reservado) > 0")
	}
	return strings.Join(condicoes, " AND "), args
}

func (repository PostgresRepository) OrdensComReservaAtiva(ctx context.Context, itemID string) ([]string, error) {
	const consulta = `SELECT DISTINCT osi.ordem_servico_id::text
		FROM reserva_estoque r
		JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id
		WHERE r.item_estoque_id = $1 AND r.status = 'ATIVA'
		ORDER BY 1`

	rows, err := repository.db.Query(ctx, consulta, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ordens []string
	for rows.Next() {
		var ordemID string
		if err := rows.Scan(&ordemID); err != nil {
			return nil, err
		}
		ordens = append(ordens, ordemID)
	}
	return ordens, rows.Err()
}

func (repository PostgresRepository) EmOrcamentoCriado(ctx context.Context, itemID string) (bool, error) {
	const consulta = `SELECT EXISTS (
		SELECT 1 FROM orcamento_item oi
		JOIN orcamento o ON o.id = oi.orcamento_id
		WHERE oi.item_estoque_id = $1 AND o.status = 'CRIADO'
	)`

	var existe bool
	err := repository.db.QueryRow(ctx, consulta, itemID).Scan(&existe)
	return existe, err
}

func (repository PostgresRepository) Desativar(ctx context.Context, item peca.Peca) error {
	const comando = `UPDATE item_estoque
		SET ativo = FALSE, data_desativacao = $2, usuario_desativacao = $3
		WHERE id = $1 AND ativo AND tipo = 'PECA'`

	etiqueta, err := repository.db.Exec(ctx, comando, item.ID, item.DataDesativacao, item.UsuarioDesativacao)
	if err != nil {
		return err
	}
	if etiqueta.RowsAffected() == 0 {
		return peca.ErrJaInativa
	}
	return nil
}

// Cadastrar insere a peca e devolve a entidade ja montada pelo BuscarPorID, que traz o
// nome da categoria e a flag de pedido em aberto sem duplicar o SELECT.
func (repository PostgresRepository) Cadastrar(ctx context.Context, cadastro peca.Cadastro) (peca.Peca, error) {
	var ativa bool
	err := repository.db.QueryRow(ctx,
		`SELECT ativa FROM categoria WHERE id = $1`, cadastro.CategoriaID).Scan(&ativa)
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}
	if err != nil {
		return peca.Peca{}, err
	}
	if !ativa {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}

	var id string
	var criadoEm time.Time
	err = repository.db.QueryRow(ctx, `
		INSERT INTO item_estoque (
			categoria_id, tipo, codigo, nome, descricao, descricao_normalizada,
			fabricante, unidade_medida, preco_venda, estoque_minimo
		) VALUES (
			$1, 'PECA', 'PEC-' || LPAD(nextval('seq_peca_codigo')::TEXT, 6, '0'),
			$2, $3, $4, $5, $6, $7, $8
		)
		RETURNING id, criado_em`,
		cadastro.CategoriaID, cadastro.Nome, cadastro.Descricao, cadastro.DescricaoNormalizada,
		cadastro.Fabricante, cadastro.UnidadeMedida, cadastro.PrecoVenda, cadastro.EstoqueMinimo,
	).Scan(&id, &criadoEm)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return peca.Peca{}, pecaApplication.ErrDescricaoDuplicada
		}
		return peca.Peca{}, err
	}

	cadastrada, err := repository.BuscarPorID(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}
	cadastrada.DataCriacao = &criadoEm
	return cadastrada, nil
}

// Atualizar aplica as alteracoes sob lock otimista, em transacao: le a linha com
// FOR UPDATE para saber o preco anterior e a version corrente, atualiza, e grava o
// historico quando o preco muda — sem que exista preco novo sem o registro.
func (repository PostgresRepository) Atualizar(ctx context.Context, id string, version int, atualizacao peca.Atualizacao, usuarioID string) (peca.Peca, error) {
	transacao, err := repository.db.Begin(ctx)
	if err != nil {
		return peca.Peca{}, err
	}
	defer func() { _ = transacao.Rollback(ctx) }()

	var ativa bool
	err = transacao.QueryRow(ctx, `SELECT ativa FROM categoria WHERE id = $1`, atualizacao.CategoriaID).Scan(&ativa)
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}
	if err != nil {
		return peca.Peca{}, err
	}
	if !ativa {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}

	var precoAnterior *string
	var versionAtual int
	err = transacao.QueryRow(ctx, `
		SELECT preco_venda::text, version FROM item_estoque
		WHERE id = $1 AND tipo = 'PECA' AND ativo
		FOR UPDATE`, id).Scan(&precoAnterior, &versionAtual)
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrNaoEncontrada
	}
	if err != nil {
		return peca.Peca{}, err
	}
	if versionAtual != version {
		return peca.Peca{}, pecaApplication.ErrVersaoDivergente
	}

	var dataAtualizacao time.Time
	var usuarioAtualizacao *string
	err = transacao.QueryRow(ctx, `
		UPDATE item_estoque
		SET nome = $2,
			descricao = $3,
			descricao_normalizada = $4,
			categoria_id = $5,
			fabricante = $6,
			preco_venda = $7::NUMERIC,
			estoque_minimo = $8,
			data_atualizacao = CURRENT_TIMESTAMP,
			usuario_atualizacao = NULLIF($9, '')::UUID,
			version = version + 1
		WHERE id = $1
		RETURNING data_atualizacao, usuario_atualizacao::text`,
		id, atualizacao.Nome, atualizacao.Descricao, atualizacao.DescricaoNormalizada,
		atualizacao.CategoriaID, atualizacao.Fabricante, atualizacao.PrecoVenda,
		atualizacao.EstoqueMinimo, usuarioID).Scan(&dataAtualizacao, &usuarioAtualizacao)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return peca.Peca{}, pecaApplication.ErrDescricaoDuplicada
		}
		return peca.Peca{}, err
	}

	if precoAnterior == nil || !mesmoPreco(*precoAnterior, atualizacao.PrecoVenda) {
		if _, err = transacao.Exec(ctx, `
			INSERT INTO historico_preco_item (item_estoque_id, preco_anterior, preco_novo, usuario_id)
			VALUES ($1, $2::NUMERIC, $3::NUMERIC, NULLIF($4, '')::UUID)`,
			id, precoAnterior, atualizacao.PrecoVenda, usuarioID); err != nil {
			return peca.Peca{}, err
		}
	}

	if err = transacao.Commit(ctx); err != nil {
		return peca.Peca{}, err
	}

	atualizada, err := repository.BuscarPorID(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}
	atualizada.DataAtualizacao = &dataAtualizacao
	atualizada.UsuarioAtualizacao = usuarioAtualizacao
	return atualizada, nil
}

// mesmoPreco compara valor, nao texto: o banco devolve "180.00" onde o cliente mandou
// "180", e um historico por diferenca de formatacao seria ruido.
func mesmoPreco(anterior, novo string) bool {
	valorAnterior, erroAnterior := strconv.ParseFloat(anterior, 64)
	valorNovo, erroNovo := strconv.ParseFloat(novo, 64)
	if erroAnterior != nil || erroNovo != nil {
		return anterior == novo
	}
	return valorAnterior == valorNovo
}
