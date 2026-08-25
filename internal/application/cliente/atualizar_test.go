package cliente

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type atualizarRepositoryFake struct {
	cliente      domain.Cliente
	buscarErr    error
	exists       bool
	existsErr    error
	atualizarErr error
	atualizado   domain.Cliente
	calls        int
}

func (fake *atualizarRepositoryFake) ExisteAtivoPorDocumento(context.Context, string) (bool, error) {
	return false, nil
}

func (fake *atualizarRepositoryFake) ExisteAtivoPorDocumentoExcetoID(context.Context, string, string) (bool, error) {
	return fake.exists, fake.existsErr
}

func (fake *atualizarRepositoryFake) BuscarPorID(context.Context, string) (domain.Cliente, error) {
	return fake.cliente, fake.buscarErr
}

func (fake *atualizarRepositoryFake) BuscarPorIDIncluindoInativo(context.Context, string) (domain.Cliente, error) {
	return domain.Cliente{}, ErrClienteNaoEncontrado
}

func (fake *atualizarRepositoryFake) BuscarPorDocumento(context.Context, string) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake *atualizarRepositoryFake) BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error) {
	return nil, nil
}

func (fake *atualizarRepositoryFake) Salvar(context.Context, domain.Cliente) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake *atualizarRepositoryFake) Atualizar(_ context.Context, cliente domain.Cliente, _ int) (domain.Cliente, error) {
	fake.calls++
	fake.atualizado = cliente
	cliente.Version++
	return cliente, fake.atualizarErr
}

func (fake *atualizarRepositoryFake) Inativar(context.Context, InativarRepositoryInput) (Inativacao, error) {
	return Inativacao{}, nil
}

func (fake *atualizarRepositoryFake) Reativar(context.Context, ReativarRepositoryInput) (Reativacao, error) {
	return Reativacao{}, nil
}

func TestAtualizarClienteUseCase(t *testing.T) {
	atual := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", TipoDocumento: domain.TipoDocumentoCPF, Telefone: "11988887777", Version: 2, Veiculos: []domain.Veiculo{{ID: "v1"}}}
	input := AtualizarInput{ClienteID: " id ", Version: 2, Dados: domain.AtualizarClienteInput{Nome: "Maria", Documento: "11222333000181", TipoDocumento: domain.TipoDocumentoCNPJ, Email: "maria@example.com"}}
	cases := []struct {
		name       string
		input      AtualizarInput
		repository *atualizarRepositoryFake
		want       error
		wantUpdate int
	}{
		{"valido", input, &atualizarRepositoryFake{cliente: atual}, nil, 1},
		{"id ausente", AtualizarInput{Version: 2}, &atualizarRepositoryFake{}, domain.ErrClienteIDObrigatorio, 0},
		{"nao encontrado", input, &atualizarRepositoryFake{buscarErr: ErrClienteNaoEncontrado}, ErrClienteNaoEncontrado, 0},
		{"versao divergente", AtualizarInput{ClienteID: "id", Version: 1, Dados: input.Dados}, &atualizarRepositoryFake{cliente: atual}, ErrVersaoDivergente, 0},
		{"dados invalidos", AtualizarInput{ClienteID: "id", Version: 2}, &atualizarRepositoryFake{cliente: atual}, domain.ErrNomeObrigatorio, 0},
		{"falha duplicidade", input, &atualizarRepositoryFake{cliente: atual, existsErr: errors.New("db")}, errors.New("db"), 0},
		{"documento duplicado", input, &atualizarRepositoryFake{cliente: atual, exists: true}, ErrClienteDuplicado, 0},
		{"falha atualizar", input, &atualizarRepositoryFake{cliente: atual, atualizarErr: errors.New("db")}, errors.New("db"), 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewAtualizar(test.repository).Execute(context.Background(), test.input)
			if test.want != nil && (err == nil || err.Error() != test.want.Error()) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || got.Version != 3 || len(test.repository.atualizado.Veiculos) != 1) {
				t.Fatalf("cliente: %#v, salvo: %#v, erro: %v", got, test.repository.atualizado, err)
			}
			if test.repository.calls != test.wantUpdate {
				t.Fatalf("atualizar chamado %d vezes", test.repository.calls)
			}
		})
	}
}
