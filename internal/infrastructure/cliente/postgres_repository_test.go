package cliente

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type fakeDB struct{ row fakeRow }

func (fake fakeDB) QueryRow(context.Context, string, ...any) pgx.Row { return fake.row }

type fakeRow struct {
	exists  bool
	cliente domain.Cliente
	err     error
}

func (row fakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if target, ok := dest[0].(*bool); ok {
		*target = row.exists
		return nil
	}
	*(dest[0].(*string)) = row.cliente.ID
	*(dest[1].(*string)) = row.cliente.Nome
	*(dest[2].(*string)) = row.cliente.Documento
	*(dest[3].(*string)) = row.cliente.TipoDocumento
	*(dest[4].(*string)) = row.cliente.Telefone
	*(dest[5].(*string)) = row.cliente.Email
	*(dest[6].(*bool)) = row.cliente.Ativo
	*(dest[7].(*int)) = row.cliente.Version
	return nil
}

func TestExisteAtivoPorDocumento(t *testing.T) {
	repository := PostgresRepository{db: fakeDB{row: fakeRow{exists: true}}}
	exists, err := repository.ExisteAtivoPorDocumento(context.Background(), "05712705402")
	if err != nil || !exists {
		t.Fatalf("exists: %v, erro: %v", exists, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).ExisteAtivoPorDocumento(context.Background(), "05712705402")
	if err == nil {
		t.Fatalf("erro: %v", err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatório")
	}
}

func TestSalvar(t *testing.T) {
	saved := domain.Cliente{ID: "id", Nome: "Teste Cliente", Documento: "52998224725", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 1}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: saved}}}).Salvar(context.Background(), domain.Cliente{})
	if err != nil || got.ID != "id" {
		t.Fatalf("cliente: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: &pgconn.PgError{Code: "23505"}}}}).Salvar(context.Background(), domain.Cliente{})
	if !errors.Is(err, application.ErrClienteDuplicado) {
		t.Fatalf("erro duplicado: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).Salvar(context.Background(), domain.Cliente{})
	if err == nil || errors.Is(err, application.ErrClienteDuplicado) {
		t.Fatalf("erro interno: %v", err)
	}
}
