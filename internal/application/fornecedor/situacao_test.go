package fornecedor

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type situacaoRepositoryStub struct {
	fornecedor domain.Fornecedor
	err        error
}

func (stub situacaoRepositoryStub) Desativar(context.Context, string, string) (domain.Fornecedor, error) {
	return stub.fornecedor, stub.err
}

func (stub situacaoRepositoryStub) Reativar(context.Context, string, string) (domain.Fornecedor, error) {
	return stub.fornecedor, stub.err
}

func TestDesativarFornecedor(t *testing.T) {
	fornecedor, err := NewDesativarFornecedor(situacaoRepositoryStub{fornecedor: domain.Fornecedor{ID: "60000000-0000-0000-0000-000000000001", Ativo: false}}).Execute(context.Background(), "60000000-0000-0000-0000-000000000001", "90000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if fornecedor.Ativo {
		t.Fatalf("fornecedor=%+v", fornecedor)
	}
}

func TestReativarFornecedor(t *testing.T) {
	fornecedor, err := NewReativarFornecedor(situacaoRepositoryStub{fornecedor: domain.Fornecedor{ID: "60000000-0000-0000-0000-000000000001", Ativo: true}}).Execute(context.Background(), "60000000-0000-0000-0000-000000000001", "90000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if !fornecedor.Ativo {
		t.Fatalf("fornecedor=%+v", fornecedor)
	}
}

func TestDesativarFornecedorRejeitaIdentificadorVazio(t *testing.T) {
	_, err := NewDesativarFornecedor(situacaoRepositoryStub{}).Execute(context.Background(), " ", "usuario")
	if !errors.Is(err, ErrSituacaoInvalida) {
		t.Fatalf("err=%v", err)
	}
}
