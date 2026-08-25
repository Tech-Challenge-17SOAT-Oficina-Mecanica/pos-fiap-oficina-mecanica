package veiculo

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

type repositoryFake struct {
	cadastrar     domain.Veiculo
	cadastrarErr  error
	consultar     domain.Veiculo
	consultarErr  error
	atualizar     domain.Veiculo
	atualizarErr  error
	veiculo       domain.Veiculo
	buscarErr     error
	ordens        []OrdemServicoAberta
	ordensErr     error
	inativacao    Inativacao
	inativarInput InativarRepositoryInput
	clienteAtivo  bool
	clienteErr    error
	placaEmUso    bool
	placaErr      error
	reativacao    Reativacao
}

func (fake *repositoryFake) CadastrarParaCliente(context.Context, string, domain.Cadastro) (domain.Veiculo, error) {
	return fake.cadastrar, fake.cadastrarErr
}
func (fake *repositoryFake) ConsultarPorPlaca(context.Context, string, bool) (domain.Veiculo, error) {
	return fake.consultar, fake.consultarErr
}
func (fake *repositoryFake) Atualizar(context.Context, string, int, domain.Cadastro) (domain.Veiculo, error) {
	return fake.atualizar, fake.atualizarErr
}
func (fake *repositoryFake) BuscarPorIDIncluindoInativo(context.Context, string) (domain.Veiculo, error) {
	return fake.veiculo, fake.buscarErr
}
func (fake *repositoryFake) BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error) {
	return fake.ordens, fake.ordensErr
}
func (fake *repositoryFake) Inativar(_ context.Context, input InativarRepositoryInput) (Inativacao, error) {
	fake.inativarInput = input
	return fake.inativacao, nil
}
func (fake *repositoryFake) ExisteAtivoPorPlacaExcetoID(context.Context, string, string) (bool, error) {
	return fake.placaEmUso, fake.placaErr
}
func (fake *repositoryFake) ClienteAtivo(context.Context, string) (bool, error) {
	return fake.clienteAtivo, fake.clienteErr
}
func (fake *repositoryFake) Reativar(context.Context, string, string) (Reativacao, error) {
	return fake.reativacao, nil
}

func TestInativar(t *testing.T) {
	ctx := context.Background()
	ativo := domain.Veiculo{ID: "v1", Ativo: true}

	t.Run("valida e inativa", func(t *testing.T) {
		fake := &repositoryFake{veiculo: ativo, inativacao: Inativacao{Veiculo: ativo}}
		if _, err := NewInativar(fake).Execute(ctx, "v1", "u1", "  duplicado  "); err != nil {
			t.Fatal(err)
		}
		if fake.inativarInput.Motivo != "duplicado" || fake.inativarInput.InativadoPor != "u1" {
			t.Fatalf("input: %+v", fake.inativarInput)
		}
	})

	for _, test := range []struct {
		name   string
		fake   *repositoryFake
		motivo string
		want   error
	}{
		{"motivo invalido", &repositoryFake{}, string(make([]byte, 201)), nil},
		{"nao encontrado", &repositoryFake{buscarErr: ErrVeiculoNaoEncontrado}, "", ErrVeiculoNaoEncontrado},
		{"ja inativo", &repositoryFake{veiculo: domain.Veiculo{Ativo: false}}, "", ErrVeiculoJaInativo},
		{"os aberta", &repositoryFake{veiculo: ativo, ordens: []OrdemServicoAberta{{OrdemServicoID: "os1"}}}, "", nil},
		{"falha ao consultar os", &repositoryFake{veiculo: ativo, ordensErr: errors.New("db")}, "", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewInativar(test.fake).Execute(ctx, "v1", "u1", test.motivo)
			if test.want != nil {
				if !errors.Is(err, test.want) {
					t.Fatalf("erro: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("erro esperado")
			}
		})
	}
}

func TestDelegarOperacoesDeVeiculo(t *testing.T) {
	ctx := context.Background()
	cadastro := domain.Cadastro{Placa: "ABC1D23"}
	fake := &repositoryFake{cadastrar: domain.Veiculo{ID: "novo"}, consultar: domain.Veiculo{ID: "consultado"}, atualizar: domain.Veiculo{ID: "atualizado"}}

	if veiculo, err := NewCadastrar(fake).Execute(ctx, "c1", cadastro); err != nil || veiculo.ID != "novo" {
		t.Fatalf("cadastrar: %+v, %v", veiculo, err)
	}
	if veiculo, err := NewConsultar(fake).Execute(ctx, "ABC1D23", true); err != nil || veiculo.ID != "consultado" {
		t.Fatalf("consultar: %+v, %v", veiculo, err)
	}
	if veiculo, err := NewAtualizar(fake).Execute(ctx, "v1", 1, cadastro); err != nil || veiculo.ID != "atualizado" {
		t.Fatalf("atualizar: %+v, %v", veiculo, err)
	}

	fake.cadastrarErr = errors.New("db")
	if _, err := NewCadastrar(fake).Execute(ctx, "c1", cadastro); err == nil {
		t.Fatal("erro de cadastro esperado")
	}
	fake.consultarErr = errors.New("db")
	if _, err := NewConsultar(fake).Execute(ctx, "ABC1D23", false); err == nil {
		t.Fatal("erro de consulta esperado")
	}
	fake.atualizarErr = errors.New("db")
	if _, err := NewAtualizar(fake).Execute(ctx, "v1", 1, cadastro); err == nil {
		t.Fatal("erro de atualizacao esperado")
	}

	if (OSAbertaError{Ordens: []OrdemServicoAberta{{OrdemServicoID: "os1"}}}).Error() == "" {
		t.Fatal("mensagem de OS aberta ausente")
	}
}

func TestReativar(t *testing.T) {
	ctx := context.Background()
	inativo := domain.Veiculo{ID: "v1", ClienteID: "c1", Cadastro: domain.Cadastro{Placa: "ABC1D23"}}

	t.Run("reativa", func(t *testing.T) {
		fake := &repositoryFake{veiculo: inativo, clienteAtivo: true, reativacao: Reativacao{Veiculo: inativo, ReativadoPor: "u1"}}
		if _, err := NewReativar(fake).Execute(ctx, "v1", "u1"); err != nil {
			t.Fatal(err)
		}
	})

	for _, test := range []struct {
		name string
		fake *repositoryFake
		want error
	}{
		{"nao encontrado", &repositoryFake{buscarErr: ErrVeiculoNaoEncontrado}, ErrVeiculoNaoEncontrado},
		{"ja ativo", &repositoryFake{veiculo: domain.Veiculo{Ativo: true}}, ErrVeiculoJaAtivo},
		{"cliente inativo", &repositoryFake{veiculo: inativo}, ErrClienteProprietarioInativo},
		{"falha ao consultar cliente", &repositoryFake{veiculo: inativo, clienteErr: errors.New("db")}, nil},
		{"placa em uso", &repositoryFake{veiculo: inativo, clienteAtivo: true, placaEmUso: true}, ErrPlacaDuplicada},
		{"falha ao consultar placa", &repositoryFake{veiculo: inativo, clienteAtivo: true, placaErr: errors.New("db")}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewReativar(test.fake).Execute(ctx, "v1", "u1")
			if test.want == nil && err != nil {
				return
			}
			if !errors.Is(err, test.want) {
				t.Fatalf("erro: %v", err)
			}
		})
	}
}
