package ordemservico

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type repositoryFake struct {
	ordem domain.OrdemDeServico
	err   error
	input CriarInput
}

func (fake *repositoryFake) Criar(_ context.Context, input CriarInput) (domain.OrdemDeServico, error) {
	fake.input = input
	return fake.ordem, fake.err
}

func TestCriarValidaIdentificadores(t *testing.T) {
	validID := "00000000-0000-0000-0000-000000000001"
	cases := []struct {
		name  string
		input CriarInput
		err   error
	}{
		{"cliente ausente", CriarInput{VeiculoID: validID}, ErrClienteIDObrigatorio},
		{"veiculo ausente", CriarInput{ClienteID: validID}, ErrVeiculoIDObrigatorio},
		{"cliente invalido", CriarInput{ClienteID: "id", VeiculoID: validID}, ErrClienteIDInvalido},
		{"veiculo invalido", CriarInput{ClienteID: validID, VeiculoID: "id"}, ErrVeiculoIDInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCriar(&repositoryFake{}).Execute(context.Background(), test.input)
			if !errors.Is(err, test.err) {
				t.Fatalf("erro = %v, esperado %v", err, test.err)
			}
		})
	}
}

func TestCriarDelegaPersistencia(t *testing.T) {
	repository := &repositoryFake{ordem: domain.OrdemDeServico{ID: "os", Status: domain.StatusRecebida}}
	input := CriarInput{ClienteID: "00000000-0000-0000-0000-000000000001", VeiculoID: "00000000-0000-0000-0000-000000000002"}
	ordem, err := NewCriar(repository).Execute(context.Background(), input)
	if err != nil || ordem.Status != domain.StatusRecebida || !reflect.DeepEqual(repository.input, input) {
		t.Fatalf("ordem = %+v, input = %+v, erro = %v", ordem, repository.input, err)
	}
}

func TestCriarComServicosEPecasNormalizaEDelega(t *testing.T) {
	repository := &repositoryFake{ordem: domain.OrdemDeServico{ID: "os", Status: domain.StatusRecebida}}
	input := CriarInput{
		ClienteID: "00000000-0000-0000-0000-000000000001",
		VeiculoID: "00000000-0000-0000-0000-000000000002",
		Servicos:  []domain.ServicoCadastro{{ServicoID: " 00000000-0000-0000-0000-000000000003 ", Observacao: " urgente "}},
		Pecas:     []ItemInput{{ItemID: "00000000-0000-0000-0000-000000000004", Quantidade: 2}},
		UsuarioID: "00000000-0000-0000-0000-000000000005",
	}
	if _, err := NewCriar(repository).Execute(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	servicos := repository.input.Servicos
	if len(servicos) != 1 || servicos[0].ServicoID != "00000000-0000-0000-0000-000000000003" || servicos[0].Observacao != "urgente" {
		t.Fatalf("servicos = %+v", servicos)
	}
	if !reflect.DeepEqual(repository.input.Pecas, input.Pecas) || repository.input.UsuarioID != input.UsuarioID {
		t.Fatalf("input = %+v", repository.input)
	}
}

func TestCriarValidaServicosEPecasAntesDePersistir(t *testing.T) {
	clienteID, veiculoID := "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"
	itemID := "00000000-0000-0000-0000-000000000003"
	cases := []struct {
		name     string
		servicos []domain.ServicoCadastro
		pecas    []ItemInput
		err      error
	}{
		{"servico invalido", []domain.ServicoCadastro{{ServicoID: "id"}}, nil, ErrServicoIDInvalido},
		{"servico duplicado", []domain.ServicoCadastro{{ServicoID: itemID}, {ServicoID: itemID}}, nil, domain.ErrServicoDuplicado},
		{"peca invalida", nil, []ItemInput{{ItemID: "id", Quantidade: 1}}, ErrItemIDInvalido},
		{"peca repetida", nil, []ItemInput{{ItemID: itemID, Quantidade: 1}, {ItemID: itemID, Quantidade: 1}}, ErrItemRepetido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repository := &repositoryFake{}
			_, err := NewCriar(repository).Execute(context.Background(), CriarInput{ClienteID: clienteID, VeiculoID: veiculoID, Servicos: test.servicos, Pecas: test.pecas})
			if !errors.Is(err, test.err) {
				t.Fatalf("erro = %v, esperado %v", err, test.err)
			}
			if repository.input.ClienteID != "" {
				t.Fatal("repository nao deveria ser chamado")
			}
		})
	}
	for _, quantidade := range []float64{0, 1.5} {
		_, err := NewCriar(&repositoryFake{}).Execute(context.Background(), CriarInput{ClienteID: clienteID, VeiculoID: veiculoID, Pecas: []ItemInput{{ItemID: itemID, Quantidade: quantidade}}})
		if err == nil {
			t.Fatalf("quantidade %v deveria ser recusada", quantidade)
		}
	}
}

func TestCriarPropagaErroDoRepository(t *testing.T) {
	expected := errors.New("db")
	input := CriarInput{ClienteID: "00000000-0000-0000-0000-000000000001", VeiculoID: "00000000-0000-0000-0000-000000000002"}
	_, err := NewCriar(&repositoryFake{err: expected}).Execute(context.Background(), input)
	if !errors.Is(err, expected) {
		t.Fatalf("erro = %v", err)
	}
}
