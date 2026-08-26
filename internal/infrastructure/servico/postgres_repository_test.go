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
	existe  bool
	servico domain.Servico
	err     error
}

func (row fakeRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if target, ok := dest[0].(*bool); ok {
		*target = row.existe
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
	*(dest[10].(*string)) = row.servico.UsuarioCriacao
	return nil
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
