package fornecedor

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type repositoryStub struct {
	fornecedor domain.Fornecedor
	err        error
}

func (stub repositoryStub) Cadastrar(context.Context, domain.Cadastro) (domain.Fornecedor, error) {
	return stub.fornecedor, stub.err
}

func TestCadastrar(t *testing.T) {
	cadastro := domain.Cadastro{Documento: "04252011000110"}
	fornecedor := domain.Fornecedor{ID: "f1"}
	resultado, err := NewCadastrar(repositoryStub{fornecedor: fornecedor}).Execute(context.Background(), cadastro)
	if err != nil || resultado.ID != "f1" {
		t.Fatalf("resultado=%+v erro=%v", resultado, err)
	}
	if _, err := NewCadastrar(repositoryStub{err: errors.New("db")}).Execute(context.Background(), cadastro); err == nil {
		t.Fatal("erro do repositório esperado")
	}
}
