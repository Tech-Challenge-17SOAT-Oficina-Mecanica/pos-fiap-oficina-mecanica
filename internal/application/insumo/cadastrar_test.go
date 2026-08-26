package insumo

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

const categoriaValida = "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724"

type cadastrarRepositoryFake struct {
	recebido insumo.Cadastro
	retorno  insumo.Insumo
	erro     error
	chamado  bool
}

func (fake *cadastrarRepositoryFake) Cadastrar(_ context.Context, cadastro insumo.Cadastro) (insumo.Insumo, error) {
	fake.chamado = true
	fake.recebido = cadastro
	return fake.retorno, fake.erro
}

func TestCadastrarInsumoDelegaAoRepositorio(t *testing.T) {
	fake := &cadastrarRepositoryFake{retorno: insumo.Insumo{ID: "id-1", Codigo: "INS-000004"}}
	cadastro := insumo.Cadastro{CategoriaID: categoriaValida, Nome: "Óleo 5W30", UnidadeMedida: "L"}

	cadastrado, err := NewCadastrarInsumo(fake).Execute(context.Background(), cadastro)
	if err != nil {
		t.Fatal(err)
	}
	if !fake.chamado {
		t.Fatal("repositório não foi chamado")
	}
	if fake.recebido.UnidadeMedida != "L" {
		t.Fatalf("cadastro repassado = %+v", fake.recebido)
	}
	if cadastrado.Codigo != "INS-000004" {
		t.Fatalf("codigo = %q", cadastrado.Codigo)
	}
}

func TestCadastrarInsumoRejeitaCategoriaNaoUUID(t *testing.T) {
	fake := &cadastrarRepositoryFake{}

	_, err := NewCadastrarInsumo(fake).Execute(context.Background(), insumo.Cadastro{CategoriaID: "nao-e-uuid"})
	if !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v, esperado %v", err, ErrIdentificadorInvalido)
	}
	if fake.chamado {
		t.Fatal("repositório não deveria ser chamado com categoria inválida")
	}
}

func TestCadastrarInsumoPropagaErro(t *testing.T) {
	fake := &cadastrarRepositoryFake{erro: ErrDescricaoDuplicada}

	_, err := NewCadastrarInsumo(fake).Execute(context.Background(), insumo.Cadastro{CategoriaID: categoriaValida})
	if !errors.Is(err, ErrDescricaoDuplicada) {
		t.Fatalf("erro = %v, esperado %v", err, ErrDescricaoDuplicada)
	}
}
