package servico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type repositoryFake struct {
	existe bool
	err    error
	salvo  domain.Servico
}

func (fake *repositoryFake) ExisteAtivoPorNomeNormalizado(context.Context, string) (bool, error) {
	return fake.existe, fake.err
}

func (fake *repositoryFake) Salvar(_ context.Context, servico domain.Servico) (domain.Servico, error) {
	fake.salvo = servico
	servico.ID = "id"
	servico.Codigo = "SER-000004"
	return servico, fake.err
}

func TestCadastrarExecute(t *testing.T) {
	input := domain.NovoServicoInput{Nome: "Revisão", Valor: "100.00", TempoEstimadoMinutos: 30}

	t.Run("sucesso", func(t *testing.T) {
		repository := &repositoryFake{}
		criado, err := NewCadastrar(repository).Execute(context.Background(), input, "usuario-id")
		if err != nil {
			t.Fatal(err)
		}
		if criado.ID == "" || criado.Codigo != "SER-000004" || repository.salvo.NomeNormalizado != "revisao" ||
			repository.salvo.UsuarioCriacao != "usuario-id" {
			t.Fatalf("serviço inesperado: %+v / salvo: %+v", criado, repository.salvo)
		}
	})

	t.Run("duplicado", func(t *testing.T) {
		_, err := NewCadastrar(&repositoryFake{existe: true}).Execute(context.Background(), input, "usuario-id")
		if !errors.Is(err, ErrServicoDuplicado) {
			t.Fatalf("erro: %v", err)
		}
	})

	t.Run("inválido", func(t *testing.T) {
		_, err := NewCadastrar(&repositoryFake{}).Execute(context.Background(), domain.NovoServicoInput{}, "usuario-id")
		if !errors.Is(err, domain.ErrNomeObrigatorio) {
			t.Fatalf("erro: %v", err)
		}
	})

	t.Run("falha no repositório", func(t *testing.T) {
		expected := errors.New("db")
		_, err := NewCadastrar(&repositoryFake{err: expected}).Execute(context.Background(), input, "usuario-id")
		if !errors.Is(err, expected) {
			t.Fatalf("erro: %v", err)
		}
	})
}
