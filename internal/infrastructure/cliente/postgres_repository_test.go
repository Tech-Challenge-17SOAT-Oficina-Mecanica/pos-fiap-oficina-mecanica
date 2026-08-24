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
	exists      bool
	cliente     domain.Cliente
	veiculosRaw []byte
	err         error
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
	if len(dest) == 9 {
		*(dest[8].(*[]byte)) = row.veiculosRaw
	}
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

func TestBuscarPorDocumento(t *testing.T) {
	cliente := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 2}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente, veiculosRaw: []byte(`[{"id":"v1","placa":"ABC1D23","marca":"Toyota","modelo":"Corolla","ano":2020}]`)}}}).BuscarPorDocumento(context.Background(), cliente.Documento)
	if err != nil || got.ID != "id" || len(got.Veiculos) != 1 {
		t.Fatalf("cliente: %#v, erro: %v", got, err)
	}
	got, err = (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente, veiculosRaw: []byte(`[]`)}}}).BuscarPorDocumento(context.Background(), cliente.Documento)
	if err != nil || len(got.Veiculos) != 0 {
		t.Fatalf("cliente sem veiculo: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).BuscarPorDocumento(context.Background(), cliente.Documento)
	if !errors.Is(err, application.ErrClienteNaoEncontrado) {
		t.Fatalf("erro nao encontrado: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).BuscarPorDocumento(context.Background(), cliente.Documento)
	if err == nil {
		t.Fatal("esperava erro")
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente, veiculosRaw: []byte(`{`)}}}).BuscarPorDocumento(context.Background(), cliente.Documento)
	if err == nil {
		t.Fatal("esperava erro de json")
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
