package peca

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
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
