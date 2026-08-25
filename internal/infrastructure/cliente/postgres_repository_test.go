package cliente

import (
	"context"
	"errors"
	"testing"
	"time"

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
	raw         []byte
	veiculosRaw []byte
	reativadoEm time.Time
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
	if len(dest) == 1 {
		*(dest[0].(*[]byte)) = row.raw
		return nil
	}
	*(dest[0].(*string)) = row.cliente.ID
	*(dest[1].(*string)) = row.cliente.Nome
	*(dest[2].(*string)) = row.cliente.Documento
	*(dest[3].(*string)) = row.cliente.TipoDocumento
	*(dest[4].(*string)) = row.cliente.Telefone
	*(dest[5].(*string)) = row.cliente.Email
	*(dest[6].(*bool)) = row.cliente.Ativo
	if target, ok := dest[7].(*int); ok {
		*target = row.cliente.Version
	}
	if target, ok := dest[7].(**time.Time); ok {
		*target = row.cliente.InativadoEm
	}
	if len(dest) >= 11 {
		*(dest[8].(*string)) = row.cliente.InativadoPor
		*(dest[9].(*string)) = row.cliente.Motivo
		*(dest[10].(*int)) = row.cliente.Version
	}
	if len(dest) == 9 {
		if target, ok := dest[8].(*[]byte); ok {
			*target = row.veiculosRaw
		}
		if target, ok := dest[8].(*time.Time); ok {
			*target = row.reativadoEm
		}
	}
	if len(dest) == 12 {
		*(dest[11].(*[]byte)) = row.veiculosRaw
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

func TestExisteAtivoPorDocumentoExcetoID(t *testing.T) {
	repository := PostgresRepository{db: fakeDB{row: fakeRow{exists: true}}}
	exists, err := repository.ExisteAtivoPorDocumentoExcetoID(context.Background(), "05712705402", "id")
	if err != nil || !exists {
		t.Fatalf("exists: %v, erro: %v", exists, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).ExisteAtivoPorDocumentoExcetoID(context.Background(), "05712705402", "id")
	if err == nil {
		t.Fatalf("erro: %v", err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatório")
	}
}

func TestBuscarPorID(t *testing.T) {
	cliente := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 2}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente}}}).BuscarPorID(context.Background(), cliente.ID)
	if err != nil || got.ID != "id" {
		t.Fatalf("cliente: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).BuscarPorID(context.Background(), cliente.ID)
	if !errors.Is(err, application.ErrClienteNaoEncontrado) {
		t.Fatalf("erro nao encontrado: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).BuscarPorID(context.Background(), cliente.ID)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestBuscarPorIDIncluindoInativo(t *testing.T) {
	now := time.Now()
	cliente := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: false, InativadoEm: &now, InativadoPor: "usuario", Motivo: "duplicado", Version: 2}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente}}}).BuscarPorIDIncluindoInativo(context.Background(), cliente.ID)
	if err != nil || got.ID != "id" || got.InativadoEm == nil || got.Motivo != "duplicado" {
		t.Fatalf("cliente: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).BuscarPorIDIncluindoInativo(context.Background(), cliente.ID)
	if !errors.Is(err, application.ErrClienteNaoEncontrado) {
		t.Fatalf("erro nao encontrado: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).BuscarPorIDIncluindoInativo(context.Background(), cliente.ID)
	if err == nil {
		t.Fatal("esperava erro")
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

func TestBuscarOSAbertas(t *testing.T) {
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{raw: []byte(`[{"ordemServicoId":"os1","status":"EM_EXECUCAO"}]`)}}}).BuscarOSAbertas(context.Background(), "id")
	if err != nil || len(got) != 1 || got[0].ID != "os1" {
		t.Fatalf("ordens: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).BuscarOSAbertas(context.Background(), "id")
	if err == nil {
		t.Fatal("esperava erro")
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{raw: []byte(`{`)}}}).BuscarOSAbertas(context.Background(), "id")
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

func TestInativar(t *testing.T) {
	now := time.Now()
	cliente := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: false, InativadoEm: &now, InativadoPor: "usuario", Motivo: "duplicado", Version: 2}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente, veiculosRaw: []byte(`[{"id":"v1","placa":"ABC1D23"}]`)}}}).Inativar(context.Background(), application.InativarRepositoryInput{ClienteID: "id", InativadoPor: "usuario", Motivo: "duplicado"})
	if err != nil || got.Cliente.Ativo || len(got.VeiculosInativados) != 1 || !got.DocumentoLiberado {
		t.Fatalf("inativacao: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).Inativar(context.Background(), application.InativarRepositoryInput{})
	if !errors.Is(err, application.ErrClienteJaInativo) {
		t.Fatalf("erro inativo: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).Inativar(context.Background(), application.InativarRepositoryInput{})
	if err == nil {
		t.Fatal("esperava erro")
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente, veiculosRaw: []byte(`{`)}}}).Inativar(context.Background(), application.InativarRepositoryInput{})
	if err == nil {
		t.Fatal("esperava erro de json")
	}
}

func TestAtualizar(t *testing.T) {
	cliente := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 3}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente}}}).Atualizar(context.Background(), cliente, 2)
	if err != nil || got.Version != 3 {
		t.Fatalf("cliente: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: &pgconn.PgError{Code: "23505"}}}}).Atualizar(context.Background(), cliente, 2)
	if !errors.Is(err, application.ErrClienteDuplicado) {
		t.Fatalf("erro duplicado: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).Atualizar(context.Background(), cliente, 2)
	if !errors.Is(err, application.ErrVersaoDivergente) {
		t.Fatalf("erro versao: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).Atualizar(context.Background(), cliente, 2)
	if err == nil || errors.Is(err, application.ErrClienteDuplicado) || errors.Is(err, application.ErrVersaoDivergente) {
		t.Fatalf("erro interno: %v", err)
	}
}

func TestReativar(t *testing.T) {
	cliente := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Ativo: true, Version: 3}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{cliente: cliente, reativadoEm: time.Now()}}}).Reativar(context.Background(), application.ReativarRepositoryInput{ClienteID: "id", ReativadoPor: "usuario"})
	if err != nil || !got.Cliente.Ativo || got.ReativadoEm.IsZero() {
		t.Fatalf("reativacao: %#v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: &pgconn.PgError{Code: "23505"}}}}).Reativar(context.Background(), application.ReativarRepositoryInput{})
	if !errors.Is(err, application.ErrClienteDuplicado) {
		t.Fatalf("erro duplicado: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).Reativar(context.Background(), application.ReativarRepositoryInput{})
	if !errors.Is(err, application.ErrClienteJaAtivo) {
		t.Fatalf("erro ativo: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).Reativar(context.Background(), application.ReativarRepositoryInput{})
	if err == nil || errors.Is(err, application.ErrClienteDuplicado) || errors.Is(err, application.ErrClienteJaAtivo) {
		t.Fatalf("erro interno: %v", err)
	}
}
