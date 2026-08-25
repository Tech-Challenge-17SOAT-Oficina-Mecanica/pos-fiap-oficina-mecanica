package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type consultarRepositoryFake struct {
	cliente domain.Cliente
	err     error
}

func (fake consultarRepositoryFake) ExisteAtivoPorDocumento(context.Context, string) (bool, error) {
	return false, nil
}

func (fake consultarRepositoryFake) ExisteAtivoPorDocumentoExcetoID(context.Context, string, string) (bool, error) {
	return false, nil
}

func (fake consultarRepositoryFake) BuscarPorID(context.Context, string) (domain.Cliente, error) {
	return domain.Cliente{}, ErrClienteNaoEncontrado
}

func (fake consultarRepositoryFake) BuscarPorIDIncluindoInativo(context.Context, string) (domain.Cliente, error) {
	return domain.Cliente{}, ErrClienteNaoEncontrado
}

func (fake consultarRepositoryFake) BuscarPorDocumento(context.Context, string) (domain.Cliente, error) {
	return fake.cliente, fake.err
}

func (fake consultarRepositoryFake) BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error) {
	return nil, nil
}

func (fake consultarRepositoryFake) Salvar(context.Context, domain.Cliente) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake consultarRepositoryFake) Atualizar(context.Context, domain.Cliente, int) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake consultarRepositoryFake) Inativar(context.Context, InativarRepositoryInput) (Inativacao, error) {
	return Inativacao{}, nil
}

func (fake consultarRepositoryFake) Reativar(context.Context, ReativarRepositoryInput) (Reativacao, error) {
	return Reativacao{}, nil
}

func TestConsultarCliente(t *testing.T) {
	cliente := domain.Cliente{ID: "id", Documento: "39053344705", Veiculos: []domain.Veiculo{{ID: "veiculo"}}}
	cases := []struct {
		name, documento string
		repository      consultarRepositoryFake
		want            error
	}{
		{"existente", "39053344705", consultarRepositoryFake{cliente: cliente}, nil},
		{"sem veiculo", "39053344705", consultarRepositoryFake{cliente: domain.Cliente{ID: "id"}}, nil},
		{"documento ausente", "", consultarRepositoryFake{}, domain.ErrDocumentoObrigatorio},
		{"documento invalido", "11111111111", consultarRepositoryFake{}, domain.ErrDocumentoInvalido},
		{"nao encontrado", "39053344705", consultarRepositoryFake{err: ErrClienteNaoEncontrado}, ErrClienteNaoEncontrado},
		{"falha repository", "39053344705", consultarRepositoryFake{err: errors.New("db")}, errors.New("db")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewConsultar(test.repository).Execute(context.Background(), test.documento)
			if test.want != nil && (err == nil || err.Error() != test.want.Error()) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || got.ID != test.repository.cliente.ID) {
				t.Fatalf("cliente: %#v, erro: %v", got, err)
			}
		})
	}
}
