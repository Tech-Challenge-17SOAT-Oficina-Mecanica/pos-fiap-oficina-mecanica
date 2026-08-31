package estoque

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

type registrarEntradaRepositoryFake struct {
	resultado Resultado
	err       error
	recebido  RegistrarEntradaInput
}

func (fake *registrarEntradaRepositoryFake) RegistrarEntrada(_ context.Context, input RegistrarEntradaInput, _ domain.EntradaCadastro) (Resultado, error) {
	fake.recebido = input
	return fake.resultado, fake.err
}

const idempotencyKeyValida = "10000000-0000-0000-0000-000000000001"
const itemIDValido = "20000000-0000-0000-0000-000000000001"

func TestRegistrarEntradaRejeitaIdempotencyKeyInvalida(t *testing.T) {
	useCase := NewRegistrarEntrada(&registrarEntradaRepositoryFake{}, nil, nil)
	_, err := useCase.Execute(context.Background(), RegistrarEntradaInput{
		IdempotencyKey: "chave-invalida", DocumentoOrigem: "NF-1", Itens: []ItemInput{{ItemID: itemIDValido, Quantidade: 1, CustoUnitario: 1}},
	})
	if err != domain.ErrIdempotencyKeyObrigatoria {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrIdempotencyKeyObrigatoria)
	}
}

func TestRegistrarEntradaRejeitaItemIDInvalido(t *testing.T) {
	useCase := NewRegistrarEntrada(&registrarEntradaRepositoryFake{}, nil, nil)
	_, err := useCase.Execute(context.Background(), RegistrarEntradaInput{
		IdempotencyKey: idempotencyKeyValida, DocumentoOrigem: "NF-1", Itens: []ItemInput{{ItemID: "item-invalido", Quantidade: 1, CustoUnitario: 1}},
	})
	if err != domain.ErrItemIDInvalido {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrItemIDInvalido)
	}
}

func TestRegistrarEntradaRejeitaFornecedorIDInvalido(t *testing.T) {
	useCase := NewRegistrarEntrada(&registrarEntradaRepositoryFake{}, nil, nil)
	_, err := useCase.Execute(context.Background(), RegistrarEntradaInput{
		IdempotencyKey: idempotencyKeyValida, DocumentoOrigem: "NF-1", FornecedorID: "fornecedor-invalido",
		Itens: []ItemInput{{ItemID: itemIDValido, Quantidade: 1, CustoUnitario: 1}},
	})
	if err != domain.ErrFornecedorIDInvalido {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrFornecedorIDInvalido)
	}
}

func TestRegistrarEntradaRejeitaSemItens(t *testing.T) {
	useCase := NewRegistrarEntrada(&registrarEntradaRepositoryFake{}, nil, nil)
	_, err := useCase.Execute(context.Background(), RegistrarEntradaInput{
		IdempotencyKey: idempotencyKeyValida, DocumentoOrigem: "NF-1",
	})
	if err != domain.ErrItensObrigatorios {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrItensObrigatorios)
	}
}

func TestRegistrarEntradaRejeitaItemRepetido(t *testing.T) {
	useCase := NewRegistrarEntrada(&registrarEntradaRepositoryFake{}, nil, nil)
	_, err := useCase.Execute(context.Background(), RegistrarEntradaInput{
		IdempotencyKey: idempotencyKeyValida, DocumentoOrigem: "NF-1",
		Itens: []ItemInput{{ItemID: itemIDValido, Quantidade: 1, CustoUnitario: 1}, {ItemID: itemIDValido, Quantidade: 2, CustoUnitario: 1}},
	})
	if err != domain.ErrItemRepetido {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrItemRepetido)
	}
}

func TestRegistrarEntradaDelegaAoRepositorio(t *testing.T) {
	fake := &registrarEntradaRepositoryFake{resultado: Resultado{Entrada: domain.ResultadoEntrada{DocumentoOrigem: "NF-1"}}}
	useCase := NewRegistrarEntrada(fake, nil, nil)
	resultado, err := useCase.Execute(context.Background(), RegistrarEntradaInput{
		IdempotencyKey: idempotencyKeyValida, DocumentoOrigem: "NF-1",
		Itens: []ItemInput{{ItemID: itemIDValido, Quantidade: 1, CustoUnitario: 1}},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Entrada.DocumentoOrigem != "NF-1" {
		t.Fatalf("documentoOrigem=%q", resultado.Entrada.DocumentoOrigem)
	}
	if fake.recebido.DocumentoOrigem != "NF-1" {
		t.Fatalf("repositorio recebeu documentoOrigem=%q", fake.recebido.DocumentoOrigem)
	}
}
