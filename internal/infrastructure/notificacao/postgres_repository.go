package notificacao

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

const colunas = `id, cliente_id, canal, tipo_evento, agregado, agregado_id,
	destinatario, assunto, conteudo, COALESCE(conteudo_html, ''), status, tentativas, ultimo_erro, criada_em, enviada_em`

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

func (repository PostgresRepository) ContatoDoCliente(ctx context.Context, clienteID string) (notificacaoApplication.Contato, error) {
	var contato notificacaoApplication.Contato
	var email *string

	err := repository.db.QueryRow(ctx,
		`SELECT id, nome, email FROM cliente WHERE id = $1 AND ativo`, clienteID).
		Scan(&contato.ID, &contato.Nome, &email)
	if errors.Is(err, pgx.ErrNoRows) {
		return notificacaoApplication.Contato{}, notificacaoApplication.ErrClienteNaoEncontrado
	}
	if err != nil {
		return notificacaoApplication.Contato{}, err
	}
	if email != nil {
		contato.Email = *email
	}
	return contato, nil
}

func (repository PostgresRepository) Enfileirar(ctx context.Context, nova notificacao.Notificacao) (notificacao.Notificacao, error) {
	err := repository.db.QueryRow(ctx, `
		INSERT INTO notificacao (
			cliente_id, canal, tipo_evento, agregado, agregado_id,
			destinatario, assunto, conteudo, conteudo_html, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), $10)
		RETURNING id, criada_em`,
		nova.ClienteID, nova.Canal, nova.TipoEvento, nova.Origem.Agregado, nova.Origem.ID,
		nova.Destinatario, nova.Assunto, nova.Conteudo, nova.ConteudoHTML, nova.Status,
	).Scan(&nova.ID, &nova.CriadaEm)
	return nova, err
}

// Pendentes traz o que ainda nao chegou ao cliente — PENDENTE e FALHOU —, das mais
// antigas primeiro. FOR UPDATE SKIP LOCKED permite mais de um processador sem que dois
// enviem a mesma notificacao.
func (repository PostgresRepository) Pendentes(ctx context.Context, limite int) ([]notificacao.Notificacao, error) {
	linhas, err := repository.db.Query(ctx, `
		SELECT `+colunas+` FROM notificacao
		WHERE status IN ('PENDENTE', 'FALHOU')
		ORDER BY criada_em
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, limite)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var pendentes []notificacao.Notificacao
	for linhas.Next() {
		aviso, err := ler(linhas)
		if err != nil {
			return nil, err
		}
		pendentes = append(pendentes, aviso)
	}
	return pendentes, linhas.Err()
}

func (repository PostgresRepository) AtualizarResultado(ctx context.Context, aviso notificacao.Notificacao) error {
	_, err := repository.db.Exec(ctx, `
		UPDATE notificacao
		SET status = $2, tentativas = $3, ultimo_erro = $4, enviada_em = $5
		WHERE id = $1`,
		aviso.ID, aviso.Status, aviso.Tentativas, aviso.UltimoErro, aviso.EnviadaEm)
	return err
}

type scanner interface {
	Scan(destinos ...any) error
}

func ler(linha scanner) (notificacao.Notificacao, error) {
	var aviso notificacao.Notificacao
	err := linha.Scan(
		&aviso.ID, &aviso.ClienteID, &aviso.Canal, &aviso.TipoEvento,
		&aviso.Origem.Agregado, &aviso.Origem.ID,
		&aviso.Destinatario, &aviso.Assunto, &aviso.Conteudo, &aviso.ConteudoHTML,
		&aviso.Status, &aviso.Tentativas, &aviso.UltimoErro,
		&aviso.CriadaEm, &aviso.EnviadaEm,
	)
	return aviso, err
}
