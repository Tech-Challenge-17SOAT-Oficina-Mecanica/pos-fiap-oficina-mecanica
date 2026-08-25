package servico

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type fakeDB struct{ row fakeRow }

func (fake fakeDB) QueryRow(context.Context, string, ...any) pgx.Row { return fake.row }

type fakeRow struct {
	existe   bool
	servico  domain.Servico
	raw      []byte
	total    int
	err      error
	situacao bool
}

func (row fakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if target, ok := dest[0].(*bool); ok {
		*target = row.existe
		return nil
	}
	if len(dest) == 2 {
		*(dest[0].(*[]byte)) = row.raw
		*(dest[1].(*int)) = row.total
		return nil
	}
	*(dest[0].(*string)) = row.servico.ID
	*(dest[1].(*string)) = row.servico.Codigo
	*(dest[2].(*string)) = row.servico.Nome
	*(dest[3].(*string)) = row.servico.NomeNormalizado
	*(dest[4].(*string)) = row.servico.Descricao
	*(dest[5].(*string)) = row.servico.Valor
	*(dest[6].(*int)) = row.servico.TempoEstimadoMinutos
	*(dest[7].(*bool)) = row.servico.Ativo
	*(dest[8].(*int)) = row.servico.Version
	*(dest[9].(*time.Time)) = row.servico.DataCriacao
	if len(dest) == 12 {
		if row.situacao {
			*(dest[10].(**time.Time)) = row.servico.DataDesativacao
			*(dest[11].(*string)) = row.servico.UsuarioDesativacao
		} else {
			*(dest[10].(**time.Time)) = row.servico.DataAtualizacao
			*(dest[11].(*string)) = row.servico.UsuarioAtualizacao
		}
	} else {
		*(dest[10].(*string)) = row.servico.UsuarioCriacao
	}
	return nil
}

func TestAlterarSituacao(t *testing.T) {
	now := time.Now()
	inativo := domain.Servico{ID: "id", Codigo: "SER-000001", Nome: "Revisão", Valor: "100.00",
		TempoEstimadoMinutos: 30, Ativo: false, Version: 2, DataCriacao: now,
		DataDesativacao: &now, UsuarioDesativacao: "usuario"}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{servico: inativo, situacao: true}}}).Desativar(context.Background(), "id", "usuario")
	if err != nil || got.Ativo || got.DataDesativacao == nil || got.UsuarioDesativacao != "usuario" {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}
	ativo := inativo
	ativo.Ativo, ativo.DataDesativacao, ativo.UsuarioDesativacao = true, nil, ""
	got, err = (PostgresRepository{db: fakeDB{row: fakeRow{servico: ativo, situacao: true}}}).Reativar(context.Background(), "id")
	if err != nil || !got.Ativo || got.DataDesativacao != nil {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).Desativar(context.Background(), "id", "usuario")
	if !errors.Is(err, domain.ErrServicoJaInativo) {
		t.Fatalf("erro: %v", err)
	}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: &pgconn.PgError{Code: "23505"}}}}).Reativar(context.Background(), "id")
	if !errors.Is(err, application.ErrServicoDuplicado) {
		t.Fatalf("erro: %v", err)
	}
}

func TestExisteAtivoPorNomeNormalizadoExcetoID(t *testing.T) {
	existe, err := (PostgresRepository{db: fakeDB{row: fakeRow{existe: true}}}).
		ExisteAtivoPorNomeNormalizadoExcetoID(context.Background(), "revisao", "id")
	if err != nil || !existe {
		t.Fatalf("existe: %v, erro: %v", existe, err)
	}
}

func TestAtualizar(t *testing.T) {
	now := time.Now()
	salvo := domain.Servico{ID: "id", Codigo: "SER-000001", Nome: "Revisão", NomeNormalizado: "revisao",
		Valor: "180.00", TempoEstimadoMinutos: 40, Ativo: true, Version: 2, DataCriacao: now,
		DataAtualizacao: &now, UsuarioAtualizacao: "usuario"}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{servico: salvo}}}).Atualizar(context.Background(), salvo, 1, "usuario")
	if err != nil || got.Version != 2 || got.DataAtualizacao == nil || got.UsuarioAtualizacao != "usuario" {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}

	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: &pgconn.PgError{Code: "23505"}}}}).
		Atualizar(context.Background(), salvo, 1, "usuario")
	if !errors.Is(err, application.ErrServicoDuplicado) {
		t.Fatalf("erro: %v", err)
	}
}

func TestListar(t *testing.T) {
	raw := []byte(`[{"id":"id","codigo":"SER-000001","nome":"Revisão","descricao":"Completa","valor":"100.00","tempoEstimadoMinutos":30,"ativo":true}]`)
	got, total, err := (PostgresRepository{db: fakeDB{row: fakeRow{raw: raw, total: 1}}}).Listar(
		context.Background(), application.Filtros{Nome: "revisão", Tamanho: 20})
	if err != nil || total != 1 || len(got) != 1 || got[0].Valor != "100.00" {
		t.Fatalf("serviços: %+v, total: %d, erro: %v", got, total, err)
	}

	_, _, err = (PostgresRepository{db: fakeDB{row: fakeRow{raw: []byte(`{`)}}}).Listar(
		context.Background(), application.Filtros{Tamanho: 20})
	if err == nil {
		t.Fatal("esperava erro de JSON")
	}

	_, _, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).Listar(
		context.Background(), application.Filtros{Tamanho: 20})
	if err == nil {
		t.Fatal("esperava erro do banco")
	}
}

func TestBuscarPorID(t *testing.T) {
	now := time.Now()
	expected := domain.Servico{ID: "id", Codigo: "SER-000001", Nome: "Revisão", NomeNormalizado: "revisao",
		Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true, Version: 2, DataCriacao: now, UsuarioCriacao: "usuario"}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{servico: expected}}}).BuscarPorID(context.Background(), "id")
	if err != nil || got.ID != expected.ID || got.Version != 2 {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}

	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: pgx.ErrNoRows}}}).BuscarPorID(context.Background(), "id")
	if !errors.Is(err, application.ErrServicoNaoEncontrado) {
		t.Fatalf("erro: %v", err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatório")
	}
}

func TestExisteAtivoPorNomeNormalizado(t *testing.T) {
	repository := PostgresRepository{db: fakeDB{row: fakeRow{existe: true}}}
	existe, err := repository.ExisteAtivoPorNomeNormalizado(context.Background(), "revisao")
	if err != nil || !existe {
		t.Fatalf("existe: %v, erro: %v", existe, err)
	}

	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).
		ExisteAtivoPorNomeNormalizado(context.Background(), "revisao")
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestSalvar(t *testing.T) {
	now := time.Now()
	salvo := domain.Servico{
		ID: "id", Codigo: "SER-000004", Nome: "Revisão", NomeNormalizado: "revisao",
		Descricao: "Completa", Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true,
		Version: 1, DataCriacao: now, UsuarioCriacao: "usuario",
	}
	got, err := (PostgresRepository{db: fakeDB{row: fakeRow{servico: salvo}}}).Salvar(context.Background(), domain.Servico{})
	if err != nil || got.ID != salvo.ID || got.Codigo != salvo.Codigo || got.DataCriacao != now {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}

	unique := &pgconn.PgError{Code: "23505"}
	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: unique}}}).Salvar(context.Background(), domain.Servico{})
	if !errors.Is(err, application.ErrServicoDuplicado) {
		t.Fatalf("erro de duplicidade: %v", err)
	}

	_, err = (PostgresRepository{db: fakeDB{row: fakeRow{err: errors.New("db")}}}).Salvar(context.Background(), domain.Servico{})
	if err == nil {
		t.Fatal("esperava erro")
	}
}
