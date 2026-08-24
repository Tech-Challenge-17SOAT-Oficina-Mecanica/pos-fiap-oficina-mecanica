package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type repositoryFake struct {
	exists      bool
	existsErr   error
	saveErr     error
	saved       domain.Cliente
	salvarCalls int
}

func (fake *repositoryFake) ExisteAtivoPorDocumento(context.Context, string) (bool, error) {
	return fake.exists, fake.existsErr
}

func (fake *repositoryFake) Salvar(_ context.Context, cliente domain.Cliente) (domain.Cliente, error) {
	fake.salvarCalls++
	fake.saved = cliente
	cliente.ID = "id"
	return cliente, fake.saveErr
}

func TestCadastrarCliente(t *testing.T) {
	cases := []struct {
		name       string
		input      domain.NovoClienteInput
		repository *repositoryFake
		want       error
		wantSave   int
	}{
		{"valido", inputValido(), &repositoryFake{}, nil, 1},
		{"dados invalidos", domain.NovoClienteInput{}, &repositoryFake{}, domain.ErrNomeObrigatorio, 0},
		{"falha ao consultar duplicidade", inputValido(), &repositoryFake{existsErr: errors.New("db")}, errors.New("db"), 0},
		{"duplicado", inputValido(), &repositoryFake{exists: true}, ErrClienteDuplicado, 0},
		{"falha ao salvar", inputValido(), &repositoryFake{saveErr: errors.New("db")}, errors.New("db"), 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewCadastrar(test.repository).Execute(context.Background(), test.input)
			if test.want != nil && (err == nil || err.Error() != test.want.Error()) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || got.ID != "id") {
				t.Fatalf("cliente: %#v, erro: %v", got, err)
			}
			if test.repository.salvarCalls != test.wantSave {
				t.Fatalf("salvar chamado %d vezes", test.repository.salvarCalls)
			}
		})
	}
}

func inputValido() domain.NovoClienteInput {
	return domain.NovoClienteInput{Nome: "Ana Martins", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777"}
}
