package orcamento

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type aprovarRepositoryFake struct {
	input     AprovarInput
	resultado domain.Aprovacao
	err       error
}

func (fake *aprovarRepositoryFake) Aprovar(_ context.Context, input AprovarInput) (domain.Aprovacao, error) {
	fake.input = input
	return fake.resultado, fake.err
}

func TestAprovarValidaEDelega(t *testing.T) {
	repository := &aprovarRepositoryFake{resultado: domain.Aprovacao{OrcamentoID: idAprovacaoValido}}
	resultado, err := NewAprovar(repository, nil, nil).Execute(context.Background(), inputAprovacaoValido())

	if err != nil || resultado.OrcamentoID != idAprovacaoValido {
		t.Fatalf("resultado=%+v erro=%v", resultado, err)
	}
	if repository.input.OrcamentoID != idAprovacaoValido {
		t.Fatalf("input nao normalizado: %+v", repository.input)
	}
}

func TestAprovarValidacoes(t *testing.T) {
	casos := []struct {
		nome  string
		mudar func(*AprovarInput)
		erro  error
	}{
		{"orcamento invalido", func(i *AprovarInput) { i.OrcamentoID = "x" }, ErrIdentificadorInvalido},
		{"decisor ausente", func(i *AprovarInput) { i.ClienteID, i.UsuarioID = "", "" }, ErrAcessoNegado},
		{"os invalida", func(i *AprovarInput) { i.OrdemServicoID = "x" }, ErrAcessoNegado},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			input := inputAprovacaoValido()
			caso.mudar(&input)
			_, err := NewAprovar(&aprovarRepositoryFake{}, nil, nil).Execute(context.Background(), input)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro=%v, esperado %v", err, caso.erro)
			}
		})
	}
}

func TestAprovarAceitaMecanico(t *testing.T) {
	repository := &aprovarRepositoryFake{}
	input := inputAprovacaoValido()
	input.ClienteID = ""
	input.UsuarioID = "30000000-0000-0000-0000-000000000001"

	if _, err := NewAprovar(repository, nil, nil).Execute(context.Background(), input); err != nil || repository.input.UsuarioID != input.UsuarioID {
		t.Fatalf("input=%+v erro=%v", repository.input, err)
	}
}

func TestAprovarPropagaErro(t *testing.T) {
	esperado := errors.New("falha")
	_, err := NewAprovar(&aprovarRepositoryFake{err: esperado}, nil, nil).Execute(context.Background(), inputAprovacaoValido())
	if !errors.Is(err, esperado) {
		t.Fatalf("erro=%v", err)
	}
}

const idAprovacaoValido = "10000000-0000-0000-0000-000000000001"

func inputAprovacaoValido() AprovarInput {
	return AprovarInput{
		OrcamentoID:    idAprovacaoValido,
		ClienteID:      "20000000-0000-0000-0000-000000000001",
		OrdemServicoID: "30000000-0000-0000-0000-000000000001",
	}
}
