package servico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type situacaoRepositoryFake struct {
	servico, salvo domain.Servico
	duplicado      bool
	err            error
	usuarioID      string
}

func (fake *situacaoRepositoryFake) BuscarPorID(context.Context, string) (domain.Servico, error) {
	return fake.servico, fake.err
}
func (fake *situacaoRepositoryFake) ExisteAtivoPorNomeNormalizadoExcetoID(context.Context, string, string) (bool, error) {
	return fake.duplicado, fake.err
}
func (fake *situacaoRepositoryFake) Desativar(_ context.Context, _ string, usuarioID string) (domain.Servico, error) {
	fake.usuarioID = usuarioID
	fake.salvo = fake.servico
	fake.salvo.Ativo = false
	return fake.salvo, fake.err
}
func (fake *situacaoRepositoryFake) Reativar(context.Context, string) (domain.Servico, error) {
	fake.salvo = fake.servico
	fake.salvo.Ativo = true
	return fake.salvo, fake.err
}

func TestDesativarExecute(t *testing.T) {
	ativo := domain.Servico{ID: "id", NomeNormalizado: "revisao", Ativo: true}
	repository := &situacaoRepositoryFake{servico: ativo}
	got, err := NewDesativar(repository).Execute(context.Background(), "id", "usuario")
	if err != nil || got.Ativo || repository.usuarioID != "usuario" {
		t.Fatalf("serviço: %+v, repo: %+v, erro: %v", got, repository, err)
	}
	_, err = NewDesativar(&situacaoRepositoryFake{servico: domain.Servico{Ativo: false}}).Execute(context.Background(), "id", "usuario")
	if !errors.Is(err, domain.ErrServicoJaInativo) {
		t.Fatalf("erro: %v", err)
	}
	_, err = NewDesativar(&situacaoRepositoryFake{err: ErrServicoNaoEncontrado}).Execute(context.Background(), "id", "usuario")
	if !errors.Is(err, ErrServicoNaoEncontrado) {
		t.Fatalf("erro: %v", err)
	}
}

func TestReativarExecute(t *testing.T) {
	inativo := domain.Servico{ID: "id", NomeNormalizado: "revisao", Ativo: false}
	got, err := NewReativar(&situacaoRepositoryFake{servico: inativo}).Execute(context.Background(), "id")
	if err != nil || !got.Ativo {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}
	_, err = NewReativar(&situacaoRepositoryFake{servico: domain.Servico{Ativo: true}}).Execute(context.Background(), "id")
	if !errors.Is(err, domain.ErrServicoJaAtivo) {
		t.Fatalf("erro: %v", err)
	}
	_, err = NewReativar(&situacaoRepositoryFake{servico: inativo, duplicado: true}).Execute(context.Background(), "id")
	if !errors.Is(err, ErrServicoDuplicado) {
		t.Fatalf("erro: %v", err)
	}
}
