package fornecedor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type dbFake struct{ row pgx.Row }

func (fake dbFake) QueryRow(context.Context, string, ...any) pgx.Row { return fake.row }

type rowFake struct {
	id      string
	ativo   bool
	version int
	criado  time.Time
	err     error
}

func (row rowFake) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	*(dest[0].(*string)) = row.id
	*(dest[1].(*bool)) = row.ativo
	*(dest[2].(*int)) = row.version
	*(dest[3].(*time.Time)) = row.criado
	return nil
}

func TestCadastrar(t *testing.T) {
	cadastro := domain.Cadastro{Documento: "04252011000110"}
	repository := PostgresRepository{db: dbFake{row: rowFake{id: "f1", ativo: true, version: 1, criado: time.Now()}}}
	fornecedor, err := repository.Cadastrar(context.Background(), cadastro)
	if err != nil || fornecedor.ID != "f1" || !fornecedor.Ativo {
		t.Fatalf("fornecedor=%+v erro=%v", fornecedor, err)
	}

	_, err = (PostgresRepository{db: dbFake{row: rowFake{err: &pgconn.PgError{Code: "23505"}}}}).Cadastrar(context.Background(), cadastro)
	if !errors.Is(err, application.ErrDocumentoDuplicado) {
		t.Fatalf("erro duplicado=%v", err)
	}
	_, err = (PostgresRepository{db: dbFake{row: rowFake{err: errors.New("db")}}}).Cadastrar(context.Background(), cadastro)
	if err == nil || errors.Is(err, application.ErrDocumentoDuplicado) {
		t.Fatalf("erro interno=%v", err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatória")
	}
}

func TestNullIfEmpty(t *testing.T) {
	if nullIfEmpty("") != nil || nullIfEmpty("valor") != "valor" {
		t.Fatal("nullIfEmpty inválido")
	}
}
