package orcamento

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type repositoryFake struct {
	total                  json.Number
	err                    error
	orcamentoID, usuarioID string
}

func (fake *repositoryFake) Calcular(_ context.Context, orcamentoID, usuarioID string) (json.Number, error) {
	fake.orcamentoID, fake.usuarioID = orcamentoID, usuarioID
	return fake.total, fake.err
}

func TestCalcular(t *testing.T) {
	repository := &repositoryFake{total: "350.00"}
	total, err := NewCalcular(repository).Execute(context.Background(), "orcamento", "usuario")
	if err != nil || total != "350.00" || repository.orcamentoID != "orcamento" || repository.usuarioID != "usuario" {
		t.Fatalf("total=%s repository=%#v erro=%v", total, repository, err)
	}
}

func TestCalcularPropagaErro(t *testing.T) {
	want := errors.New("db")
	_, err := NewCalcular(&repositoryFake{err: want}).Execute(context.Background(), "orcamento", "usuario")
	if !errors.Is(err, want) {
		t.Fatalf("erro=%v", err)
	}
}
