package notificacao

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	app "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	infra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/notificacao"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func bancoIsolado(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	root, err := database.OpenPool()
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("notificacao_teste_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	_, err = root.Exec(ctx, "CREATE SCHEMA "+quoted+"; CREATE TABLE "+quoted+".cliente (LIKE public.cliente INCLUDING ALL); CREATE TABLE "+quoted+".notificacao (LIKE public.notificacao INCLUDING ALL)")
	if err != nil {
		root.Close()
		t.Fatal(err)
	}
	config := root.Config()
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close(); _, _ = root.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); root.Close() })
	return pool
}

func enfileirarTeste(t *testing.T, db *pgxpool.Pool) domain.Notificacao {
	t.Helper()
	ctx := context.Background()
	const id = "20000000-0000-0000-1000-000000000001"
	_, err := db.Exec(ctx, `INSERT INTO cliente (id,nome,documento,tipo_documento,telefone,email,ativo) VALUES ($1,'Teste','12345678909','CPF','11999999999','teste@example.com',true) ON CONFLICT DO NOTHING`, id)
	if err != nil {
		t.Fatal(err)
	}
	n, err := domain.Nova(id, "teste@example.com", domain.EventoServicoFinalizado, "Teste", "Sem envio externo", domain.Origem{Agregado: "teste", ID: id})
	if err != nil {
		t.Fatal(err)
	}
	n, err = infra.NewPostgresRepository(db).Enfileirar(ctx, n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

type enviadorFunc func(context.Context, domain.Notificacao) error

func (fn enviadorFunc) Enviar(ctx context.Context, n domain.Notificacao) error { return fn(ctx, n) }

func TestDoisConsumidoresNaoEnviamMesmaNotificacao(t *testing.T) {
	db := bancoIsolado(t)
	n := enfileirarTeste(t, db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	entered, release := make(chan struct{}), make(chan struct{})
	var sends atomic.Int32
	first := enviadorFunc(func(ctx context.Context, n domain.Notificacao) error {
		sends.Add(1)
		close(entered)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	done := make(chan error, 1)
	go func() {
		_, err := app.NewProcessar(infra.NewPostgresRepository(db), first).Execute(ctx, 1)
		done <- err
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("primeiro consumidor nao iniciou")
	}
	// O primeiro permanece no envio: o segundo deve pular a linha reservada.
	second := enviadorFunc(func(context.Context, domain.Notificacao) error { sends.Add(1); return nil })
	result, err := app.NewProcessar(infra.NewPostgresRepository(db), second).Execute(ctx, 1)
	close(release)
	if err != nil {
		t.Fatal(err)
	}
	if result.Processadas != 0 {
		t.Errorf("segundo consumidor processou linha reservada: %+v", result)
	}
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if sends.Load() != 1 {
		t.Fatalf("envios=%d, esperado 1", sends.Load())
	}
	var attempts int
	var state string
	if err = db.QueryRow(ctx, "SELECT tentativas,status FROM notificacao WHERE id=$1", n.ID).Scan(&attempts, &state); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 || state != "ENVIADA" {
		t.Fatalf("tentativas=%d status=%s", attempts, state)
	}
}

func TestCancelamentoLiberaReservaParaOutraTentativa(t *testing.T) {
	db := bancoIsolado(t)
	n := enfileirarTeste(t, db)
	repo := infra.NewPostgresRepository(db)
	err := repo.ProcessarPendentes(context.Background(), 1, func(context.Context, domain.Notificacao) (domain.Notificacao, error) {
		return domain.Notificacao{}, context.Canceled
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("erro=%v", err)
	}
	result, err := app.NewProcessar(repo, enviadorFunc(func(context.Context, domain.Notificacao) error { return nil })).Execute(context.Background(), 1)
	if err != nil || result.Enviadas != 1 {
		t.Fatalf("reserva nao liberada: %+v %v", result, err)
	}
	var attempts int
	if err = db.QueryRow(context.Background(), "SELECT tentativas FROM notificacao WHERE id=$1", n.ID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 {
		t.Fatalf("tentativas=%d", attempts)
	}
}

func TestFalhaNaoRetentaMesmaMensagemNaMesmaRodada(t *testing.T) {
	db := bancoIsolado(t)
	enfileirarTeste(t, db)
	repo := infra.NewPostgresRepository(db)
	var calls int
	sender := enviadorFunc(func(context.Context, domain.Notificacao) error { calls++; return errors.New("falha simulada") })
	result, err := app.NewProcessar(repo, sender).Execute(context.Background(), 20)
	if err != nil || calls != 1 || result.Falhas != 1 {
		t.Fatalf("chamadas=%d resultado=%+v erro=%v", calls, result, err)
	}
	_, err = app.NewProcessar(repo, sender).Execute(context.Background(), 20)
	if err != nil || calls != 2 {
		t.Fatalf("nova rodada deveria retentar uma vez: %d %v", calls, err)
	}
}
