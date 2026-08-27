package fornecedor

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type atualizacaoRepositoryStub struct {
	fornecedor  domain.Fornecedor
	atualizacao domain.Atualizacao
	version     int
	usuarioID   string
	err         error
}

func (stub *atualizacaoRepositoryStub) Atualizar(_ context.Context, _ string, atualizacao domain.Atualizacao, version int, usuarioID string) (domain.Fornecedor, error) {
	stub.atualizacao = atualizacao
	stub.version = version
	stub.usuarioID = usuarioID
	return stub.fornecedor, stub.err
}

func TestAtualizarFornecedor(t *testing.T) {
	repository := &atualizacaoRepositoryStub{fornecedor: domain.Fornecedor{ID: "60000000-0000-0000-0000-000000000001", Version: 2}}
	atualizacao, err := domain.NovaAtualizacao("Auto Pecas Brasil LTDA", "Auto Pecas", "11999991001", "vendas@example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	fornecedor, err := NewAtualizarFornecedor(repository).Execute(context.Background(), "60000000-0000-0000-0000-000000000001", atualizacao, 1, "90000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if fornecedor.Version != 2 || repository.version != 1 || repository.usuarioID == "" {
		t.Fatalf("fornecedor=%+v repository=%+v", fornecedor, repository)
	}
}

func TestAtualizarFornecedorRejeitaVersionInvalida(t *testing.T) {
	_, err := NewAtualizarFornecedor(&atualizacaoRepositoryStub{}).Execute(context.Background(), "60000000-0000-0000-0000-000000000001", domain.Atualizacao{}, 0, "usuario")
	if !errors.Is(err, ErrAtualizacaoInvalida) {
		t.Fatalf("err=%v", err)
	}
}
