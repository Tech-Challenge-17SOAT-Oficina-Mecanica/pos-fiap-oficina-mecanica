package cliente

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type deletarRepositoryFake struct {
	cliente    domain.Cliente
	buscarErr  error
	osAbertas  []OrdemServicoAberta
	osErr      error
	exists     bool
	existsErr  error
	inativacao Inativacao
	reativacao Reativacao
	err        error
}

func (fake *deletarRepositoryFake) ExisteAtivoPorDocumento(context.Context, string) (bool, error) {
	return false, nil
}

func (fake *deletarRepositoryFake) ExisteAtivoPorDocumentoExcetoID(context.Context, string, string) (bool, error) {
	return fake.exists, fake.existsErr
}

func (fake *deletarRepositoryFake) BuscarPorID(context.Context, string) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake *deletarRepositoryFake) BuscarPorIDIncluindoInativo(context.Context, string) (domain.Cliente, error) {
	return fake.cliente, fake.buscarErr
}

func (fake *deletarRepositoryFake) BuscarPorDocumento(context.Context, string) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake *deletarRepositoryFake) BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error) {
	return fake.osAbertas, fake.osErr
}

func (fake *deletarRepositoryFake) Salvar(context.Context, domain.Cliente) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake *deletarRepositoryFake) Atualizar(context.Context, domain.Cliente, int) (domain.Cliente, error) {
	return domain.Cliente{}, nil
}

func (fake *deletarRepositoryFake) Inativar(context.Context, InativarRepositoryInput) (Inativacao, error) {
	return fake.inativacao, fake.err
}

func (fake *deletarRepositoryFake) Reativar(context.Context, ReativarRepositoryInput) (Reativacao, error) {
	return fake.reativacao, fake.err
}

func TestInativarCliente(t *testing.T) {
	now := time.Now()
	clienteAtivo := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", Ativo: true}
	inativacao := Inativacao{Cliente: domain.Cliente{ID: "id", Nome: "Ana", Ativo: false, InativadoEm: &now, InativadoPor: "usuario"}, DocumentoLiberado: true}
	cases := []struct {
		name       string
		input      InativarInput
		repository *deletarRepositoryFake
		want       error
	}{
		{"valido", InativarInput{ClienteID: " id ", UsuarioID: "usuario", Motivo: " duplicado "}, &deletarRepositoryFake{cliente: clienteAtivo, inativacao: inativacao}, nil},
		{"id ausente", InativarInput{}, &deletarRepositoryFake{}, domain.ErrClienteIDObrigatorio},
		{"motivo invalido", InativarInput{ClienteID: "id", Motivo: string(make([]byte, 201))}, &deletarRepositoryFake{}, domain.ErrMotivoInvalido},
		{"nao encontrado", InativarInput{ClienteID: "id"}, &deletarRepositoryFake{buscarErr: ErrClienteNaoEncontrado}, ErrClienteNaoEncontrado},
		{"ja inativo", InativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: domain.Cliente{ID: "id", Ativo: false}}, ErrClienteJaInativo},
		{"falha os", InativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: clienteAtivo, osErr: errors.New("db")}, errors.New("db")},
		{"os aberta", InativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: clienteAtivo, osAbertas: []OrdemServicoAberta{{ID: "os", Status: "EM_EXECUCAO"}}}, ErrClienteComOSAberta},
		{"falha inativar", InativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: clienteAtivo, err: errors.New("db")}, errors.New("db")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewInativar(test.repository).Execute(context.Background(), test.input)
			if test.want != nil && (err == nil || (!errors.Is(err, test.want) && err.Error() != test.want.Error())) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || got.Cliente.Ativo) {
				t.Fatalf("inativacao: %#v, erro: %v", got, err)
			}
		})
	}
}

func TestReativarCliente(t *testing.T) {
	clienteInativo := domain.Cliente{ID: "id", Nome: "Ana", Documento: "39053344705", Ativo: false}
	reativacao := Reativacao{Cliente: domain.Cliente{ID: "id", Nome: "Ana", Ativo: true}, ReativadoEm: time.Now()}
	cases := []struct {
		name       string
		input      ReativarInput
		repository *deletarRepositoryFake
		want       error
	}{
		{"valido", ReativarInput{ClienteID: " id ", UsuarioID: "usuario"}, &deletarRepositoryFake{cliente: clienteInativo, reativacao: reativacao}, nil},
		{"id ausente", ReativarInput{}, &deletarRepositoryFake{}, domain.ErrClienteIDObrigatorio},
		{"nao encontrado", ReativarInput{ClienteID: "id"}, &deletarRepositoryFake{buscarErr: ErrClienteNaoEncontrado}, ErrClienteNaoEncontrado},
		{"ja ativo", ReativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: domain.Cliente{ID: "id", Ativo: true}}, ErrClienteJaAtivo},
		{"falha duplicidade", ReativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: clienteInativo, existsErr: errors.New("db")}, errors.New("db")},
		{"duplicado", ReativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: clienteInativo, exists: true}, ErrClienteDuplicado},
		{"falha reativar", ReativarInput{ClienteID: "id"}, &deletarRepositoryFake{cliente: clienteInativo, err: errors.New("db")}, errors.New("db")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewReativar(test.repository).Execute(context.Background(), test.input)
			if test.want != nil && (err == nil || (!errors.Is(err, test.want) && err.Error() != test.want.Error())) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || !got.Cliente.Ativo) {
				t.Fatalf("reativacao: %#v, erro: %v", got, err)
			}
		})
	}
}

func TestOSAbertaError(t *testing.T) {
	err := OSAbertaError{Ordens: []OrdemServicoAberta{{ID: "os", Status: "EM_EXECUCAO"}}}
	if err.Error() != ErrClienteComOSAberta.Error() || !errors.Is(err, ErrClienteComOSAberta) {
		t.Fatalf("erro: %v", err)
	}
}
