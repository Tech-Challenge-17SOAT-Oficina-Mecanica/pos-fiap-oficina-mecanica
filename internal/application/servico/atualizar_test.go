package servico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type atualizarRepositoryFake struct {
	servico    domain.Servico
	duplicado  bool
	err        error
	atualizado domain.Servico
	version    int
	usuarioID  string
}

func (fake *atualizarRepositoryFake) BuscarPorID(context.Context, string) (domain.Servico, error) {
	return fake.servico, fake.err
}
func (fake *atualizarRepositoryFake) ExisteAtivoPorNomeNormalizadoExcetoID(context.Context, string, string) (bool, error) {
	return fake.duplicado, fake.err
}
func (fake *atualizarRepositoryFake) Atualizar(_ context.Context, servico domain.Servico, version int, usuarioID string) (domain.Servico, error) {
	fake.atualizado, fake.version, fake.usuarioID = servico, version, usuarioID
	servico.Version++
	return servico, fake.err
}

func TestAtualizarExecute(t *testing.T) {
	original := domain.Servico{ID: "id", Codigo: "SER-000001", Nome: "Revisão", NomeNormalizado: "revisao",
		Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true, Version: 2}
	nome := "Revisão completa"

	t.Run("sucesso", func(t *testing.T) {
		repository := &atualizarRepositoryFake{servico: original}
		got, err := NewAtualizar(repository).Execute(context.Background(), "id", 2, domain.Atualizacao{Nome: &nome}, "usuario")
		if err != nil || got.Version != 3 || repository.atualizado.Codigo != original.Codigo || repository.version != 2 || repository.usuarioID != "usuario" {
			t.Fatalf("serviço: %+v, repo: %+v, erro: %v", got, repository, err)
		}
	})

	t.Run("versão divergente", func(t *testing.T) {
		_, err := NewAtualizar(&atualizarRepositoryFake{servico: original}).Execute(context.Background(), "id", 1, domain.Atualizacao{Nome: &nome}, "usuario")
		if !errors.Is(err, ErrVersaoDivergente) {
			t.Fatalf("erro: %v", err)
		}
	})

	t.Run("duplicidade", func(t *testing.T) {
		_, err := NewAtualizar(&atualizarRepositoryFake{servico: original, duplicado: true}).Execute(context.Background(), "id", 2, domain.Atualizacao{Nome: &nome}, "usuario")
		if !errors.Is(err, ErrServicoDuplicado) {
			t.Fatalf("erro: %v", err)
		}
	})

	t.Run("erro ao buscar", func(t *testing.T) {
		expected := errors.New("db")
		_, err := NewAtualizar(&atualizarRepositoryFake{err: expected}).Execute(context.Background(), "id", 2, domain.Atualizacao{Nome: &nome}, "usuario")
		if !errors.Is(err, expected) {
			t.Fatalf("erro: %v", err)
		}
	})
}
