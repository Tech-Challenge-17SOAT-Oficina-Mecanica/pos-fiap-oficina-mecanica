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
	resultado, err := NewAprovar(repository).Execute(context.Background(), inputAprovacaoValido())

	if err != nil || resultado.OrcamentoID != idAprovacaoValido {
		t.Fatalf("resultado=%+v erro=%v", resultado, err)
	}
	if repository.input.FornecedorID != "40000000-0000-0000-0000-000000000001" {
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
		{"cliente invalido", func(i *AprovarInput) { i.ClienteID = "" }, ErrAcessoNegado},
		{"os invalida", func(i *AprovarInput) { i.OrdemServicoID = "x" }, ErrAcessoNegado},
		{"fornecedor invalido", func(i *AprovarInput) { i.FornecedorID = "x" }, ErrFornecedorIDInvalido},
	}
	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			input := inputAprovacaoValido()
			caso.mudar(&input)
			_, err := NewAprovar(&aprovarRepositoryFake{}).Execute(context.Background(), input)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro=%v, esperado %v", err, caso.erro)
			}
		})
	}
}

func TestAprovarPropagaErro(t *testing.T) {
	esperado := errors.New("falha")
	_, err := NewAprovar(&aprovarRepositoryFake{err: esperado}).Execute(context.Background(), inputAprovacaoValido())
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
		FornecedorID:   "40000000-0000-0000-0000-000000000001",
	}
}
