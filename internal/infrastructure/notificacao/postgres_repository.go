package notificacao

import (
	"context"
	"errors"
	"time"

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

// ProcessarPendentes abre uma transacao por mensagem. O lock so e liberado
// depois do envio e do registro do resultado; outro consumidor usa SKIP LOCKED.
// IDs ja tratados nao sao retentados na mesma rodada, inclusive os que falharam.
func (repository PostgresRepository) ProcessarPendentes(ctx context.Context, limite int, tratar func(context.Context, notificacao.Notificacao) (notificacao.Notificacao, error)) error {
	ids := make([]string, 0)
	for i := 0; i < limite; i++ {
		id, ok, err := repository.processarProxima(ctx, ids, tratar)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		ids = append(ids, id)
	}
	return nil
}

func (repository PostgresRepository) processarProxima(ctx context.Context, ignorar []string, tratar func(context.Context, notificacao.Notificacao) (notificacao.Notificacao, error)) (string, bool, error) {
	ctx, cancelar := context.WithTimeout(ctx, 15*time.Second)
	defer cancelar()
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback(context.Background())
	aviso, err := ler(tx.QueryRow(ctx, `
		SELECT `+colunas+` FROM notificacao
		WHERE status IN ('PENDENTE', 'FALHOU') AND NOT (id = ANY($1::uuid[]))
		ORDER BY criada_em, id
		LIMIT 1 FOR UPDATE SKIP LOCKED`, ignorar))
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	atualizada, err := tratar(ctx, aviso)
	if err != nil {
		return "", false, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE notificacao SET status=$2, tentativas=$3, ultimo_erro=$4, enviada_em=$5
		WHERE id=$1`, aviso.ID, atualizada.Status, atualizada.Tentativas, atualizada.UltimoErro, atualizada.EnviadaEm)
	if err != nil {
		return "", false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", false, err
	}
	return aviso.ID, true, nil
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
