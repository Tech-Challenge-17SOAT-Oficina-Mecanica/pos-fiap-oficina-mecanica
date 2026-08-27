package ordemservico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type repositoryFake struct {
	ordem domain.OrdemDeServico
	err   error
	input CriarInput
}

func (fake *repositoryFake) Criar(_ context.Context, input CriarInput) (domain.OrdemDeServico, error) {
	fake.input = input
	return fake.ordem, fake.err
}

func TestCriarValidaIdentificadores(t *testing.T) {
	validID := "00000000-0000-0000-0000-000000000001"
	cases := []struct {
		name  string
		input CriarInput
		err   error
	}{
		{"cliente ausente", CriarInput{VeiculoID: validID}, ErrClienteIDObrigatorio},
		{"veiculo ausente", CriarInput{ClienteID: validID}, ErrVeiculoIDObrigatorio},
		{"cliente invalido", CriarInput{ClienteID: "id", VeiculoID: validID}, ErrClienteIDInvalido},
		{"veiculo invalido", CriarInput{ClienteID: validID, VeiculoID: "id"}, ErrVeiculoIDInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCriar(&repositoryFake{}).Execute(context.Background(), test.input)
			if !errors.Is(err, test.err) {
				t.Fatalf("erro = %v, esperado %v", err, test.err)
			}
		})
	}
}

func TestCriarDelegaPersistencia(t *testing.T) {
	repository := &repositoryFake{ordem: domain.OrdemDeServico{ID: "os", Status: domain.StatusRecebida}}
	input := CriarInput{ClienteID: "00000000-0000-0000-0000-000000000001", VeiculoID: "00000000-0000-0000-0000-000000000002"}
	ordem, err := NewCriar(repository).Execute(context.Background(), input)
	if err != nil || ordem.Status != domain.StatusRecebida || repository.input != input {
		t.Fatalf("ordem = %+v, input = %+v, erro = %v", ordem, repository.input, err)
	}
}

func TestCriarPropagaErroDoRepository(t *testing.T) {
	expected := errors.New("db")
	input := CriarInput{ClienteID: "00000000-0000-0000-0000-000000000001", VeiculoID: "00000000-0000-0000-0000-000000000002"}
	_, err := NewCriar(&repositoryFake{err: expected}).Execute(context.Background(), input)
	if !errors.Is(err, expected) {
		t.Fatalf("erro = %v", err)
	}
}
