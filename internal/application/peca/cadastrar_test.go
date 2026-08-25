package peca

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type cadastrarRepositoryFake struct {
	recebido peca.Cadastro
	retorno  peca.Peca
	erro     error
	chamado  bool
}

func (fake *cadastrarRepositoryFake) Cadastrar(_ context.Context, cadastro peca.Cadastro) (peca.Peca, error) {
	fake.chamado = true
	fake.recebido = cadastro
	return fake.retorno, fake.erro
}

func TestCadastrarPecaDelegaAoRepositorio(t *testing.T) {
	fake := &cadastrarRepositoryFake{retorno: peca.Peca{ID: "id-1", Codigo: "PEC-000003"}}
	cadastro := peca.Cadastro{CategoriaID: "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94", Nome: "Pastilha"}

	cadastrada, err := NewCadastrarPeca(fake).Execute(context.Background(), cadastro)
	if err != nil {
		t.Fatal(err)
	}
	if !fake.chamado {
		t.Fatal("repositório não foi chamado")
	}
	if fake.recebido.CategoriaID != cadastro.CategoriaID {
		t.Fatalf("cadastro repassado = %+v", fake.recebido)
	}
	if cadastrada.Codigo != "PEC-000003" {
		t.Fatalf("codigo = %q", cadastrada.Codigo)
	}
}

func TestCadastrarPecaRejeitaCategoriaNaoUUID(t *testing.T) {
	fake := &cadastrarRepositoryFake{}

	_, err := NewCadastrarPeca(fake).Execute(context.Background(), peca.Cadastro{CategoriaID: "nao-e-uuid"})
	if !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v, esperado %v", err, ErrIdentificadorInvalido)
	}
	if fake.chamado {
		t.Fatal("repositório não deveria ser chamado com categoria inválida")
	}
}

func TestCadastrarPecaPropagaErro(t *testing.T) {
	fake := &cadastrarRepositoryFake{erro: ErrDescricaoDuplicada}

	_, err := NewCadastrarPeca(fake).Execute(context.Background(),
		peca.Cadastro{CategoriaID: "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94"})
	if !errors.Is(err, ErrDescricaoDuplicada) {
		t.Fatalf("erro = %v, esperado %v", err, ErrDescricaoDuplicada)
	}
}
