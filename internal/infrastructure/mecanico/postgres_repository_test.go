package mecanico

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

type fakeDB struct {
	row fakeRow
}

func (fake fakeDB) QueryRow(context.Context, string, ...any) pgx.Row { return fake.row }

type fakeRow struct {
	exists   bool
	mecanico domain.Mecanico
	err      error
}

func (row fakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if target, ok := dest[0].(*bool); ok {
		*target = row.exists
		return nil
	}
	if target, ok := dest[2].(*bool); ok {
		*(dest[0].(*string)) = row.mecanico.UsuarioID
		*(dest[1].(*string)) = row.mecanico.Email
		*target = row.mecanico.Ativo
		return nil
	}
	*(dest[0].(*string)) = row.mecanico.ID
	*(dest[1].(*string)) = row.mecanico.Nome
	*(dest[2].(*int)) = row.mecanico.Version
	return nil
}

type fakeTx struct {
	rows      []fakeRow
	execErr   error
	commitErr error
	queries   int
	execs     int
	commits   int
	rollbacks int
}

func (fake *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row {
	row := fake.rows[fake.queries]
	fake.queries++
	return row
}

func (fake *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	fake.execs++
	return pgconn.CommandTag{}, fake.execErr
}

func (fake *fakeTx) Commit(context.Context) error {
	fake.commits++
	return fake.commitErr
}

func (fake *fakeTx) Rollback(context.Context) error {
	fake.rollbacks++
	return nil
}

func TestNewPostgresRepository(t *testing.T) {
	repository := NewPostgresRepository(nil)
	if repository.begin == nil {
		t.Fatal("begin obrigatório")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("esperava panic ao abrir transação sem db")
			}
		}()
		_, _ = repository.begin(context.Background())
	}()

	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(context.Background()); err != nil {
		t.Skip("banco indisponível")
	}
	repository = NewPostgresRepository(db)
	if repository.db == nil || repository.begin == nil {
		t.Fatal("db obrigatório")
	}
	tx, err := repository.begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestEmailExiste(t *testing.T) {
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{exists: true}}}).EmailExiste(context.Background(), "m@oficina.local")
	if err != nil || !got {
		t.Fatalf("exists: %v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).EmailExiste(context.Background(), "m@oficina.local")
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestSalvarMecanico(t *testing.T) {
	mecanico := domain.Mecanico{Nome: "Maria", Email: "maria@oficina.local", Escopos: []string{"clientes:ler"}}
	t.Run("sucesso", func(t *testing.T) {
		tx := &fakeTx{rows: []fakeRow{
			{mecanico: domain.Mecanico{UsuarioID: "u1", Email: "maria@oficina.local", Ativo: true}},
			{mecanico: domain.Mecanico{ID: "m1", Nome: "Maria", Version: 1}},
		}}
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return tx, nil }}
		got, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if err != nil || got.ID != "m1" || got.Version != 1 || tx.execs != 1 || tx.commits != 1 || tx.rollbacks != 1 {
			t.Fatalf("mecanico: %#v, tx: %#v, erro: %v", got, tx, err)
		}
	})
	t.Run("erro ao abrir transacao", func(t *testing.T) {
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return nil, errors.New("begin") }}
		_, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if err == nil {
			t.Fatal("esperava erro")
		}
	})
	t.Run("email duplicado", func(t *testing.T) {
		tx := &fakeTx{rows: []fakeRow{{err: &pgconn.PgError{Code: "23505"}}}}
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return tx, nil }}
		_, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if !errors.Is(err, application.ErrEmailDuplicado) || tx.rollbacks != 1 {
			t.Fatalf("erro: %v, rollbacks: %d", err, tx.rollbacks)
		}
	})
	t.Run("erro ao criar mecanico", func(t *testing.T) {
		tx := &fakeTx{rows: []fakeRow{
			{mecanico: domain.Mecanico{UsuarioID: "u1", Email: "maria@oficina.local", Ativo: true}},
			{err: errors.New("db")},
		}}
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return tx, nil }}
		_, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if err == nil || tx.rollbacks != 1 {
			t.Fatalf("erro: %v, rollbacks: %d", err, tx.rollbacks)
		}
	})
	t.Run("erro ao criar escopo", func(t *testing.T) {
		tx := &fakeTx{execErr: errors.New("db"), rows: []fakeRow{
			{mecanico: domain.Mecanico{UsuarioID: "u1", Email: "maria@oficina.local", Ativo: true}},
			{mecanico: domain.Mecanico{ID: "m1", Nome: "Maria", Version: 1}},
		}}
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return tx, nil }}
		_, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if err == nil || tx.rollbacks != 1 {
			t.Fatalf("erro: %v, rollbacks: %d", err, tx.rollbacks)
		}
	})
	t.Run("erro ao criar usuario", func(t *testing.T) {
		tx := &fakeTx{rows: []fakeRow{{err: errors.New("db")}}}
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return tx, nil }}
		_, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if err == nil || tx.rollbacks != 1 {
			t.Fatalf("erro: %v, rollbacks: %d", err, tx.rollbacks)
		}
	})
	t.Run("erro no commit", func(t *testing.T) {
		tx := &fakeTx{commitErr: errors.New("commit"), rows: []fakeRow{
			{mecanico: domain.Mecanico{UsuarioID: "u1", Email: "maria@oficina.local", Ativo: true}},
			{mecanico: domain.Mecanico{ID: "m1", Nome: "Maria", Version: 1}},
		}}
		repository := PostgresRepository{begin: func(context.Context) (transaction, error) { return tx, nil }}
		_, err := repository.SalvarMecanico(context.Background(), mecanico, "hash")
		if err == nil || tx.commits != 1 {
			t.Fatalf("erro: %v, commits: %d", err, tx.commits)
		}
	})
}
