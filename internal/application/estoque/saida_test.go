package estoque

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

type registrarSaidaRepositoryFake struct {
	resultado ResultadoSaida
	err       error
	recebido  RegistrarSaidaInput
}

func (fake *registrarSaidaRepositoryFake) RegistrarSaida(_ context.Context, input RegistrarSaidaInput, _ domain.SaidaCadastro) (ResultadoSaida, error) {
	fake.recebido = input
	return fake.resultado, fake.err
}

const ordemServicoIDValida = "30000000-0000-0000-0000-000000000001"

func TestRegistrarSaidaRejeitaIdempotencyKeyInvalida(t *testing.T) {
	useCase := NewRegistrarSaida(&registrarSaidaRepositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarSaidaInput{
		IdempotencyKey: "chave-invalida", OrdemServicoID: ordemServicoIDValida,
		Itens: []ItemSaidaInput{{ItemID: itemIDValido, Quantidade: 1}},
	})
	if err != domain.ErrIdempotencyKeyObrigatoria {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrIdempotencyKeyObrigatoria)
	}
}

func TestRegistrarSaidaRejeitaOrdemServicoIDInvalida(t *testing.T) {
	useCase := NewRegistrarSaida(&registrarSaidaRepositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarSaidaInput{
		IdempotencyKey: idempotencyKeyValida, OrdemServicoID: "os-invalida",
		Itens: []ItemSaidaInput{{ItemID: itemIDValido, Quantidade: 1}},
	})
	if err != ErrOrdemServicoNaoEncontrada {
		t.Fatalf("erro=%v, esperado %v", err, ErrOrdemServicoNaoEncontrada)
	}
}

func TestRegistrarSaidaRejeitaItemIDInvalido(t *testing.T) {
	useCase := NewRegistrarSaida(&registrarSaidaRepositoryFake{})
	_, err := useCase.Execute(context.Background(), RegistrarSaidaInput{
		IdempotencyKey: idempotencyKeyValida, OrdemServicoID: ordemServicoIDValida,
		Itens: []ItemSaidaInput{{ItemID: "item-invalido", Quantidade: 1}},
	})
	if err != domain.ErrItemIDInvalido {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrItemIDInvalido)
	}
}

func TestRegistrarSaidaDelegaAoRepositorio(t *testing.T) {
	fake := &registrarSaidaRepositoryFake{resultado: ResultadoSaida{Saida: domain.ResultadoSaida{OrdemServicoID: ordemServicoIDValida}}}
	useCase := NewRegistrarSaida(fake)
	resultado, err := useCase.Execute(context.Background(), RegistrarSaidaInput{
		IdempotencyKey: idempotencyKeyValida, OrdemServicoID: ordemServicoIDValida,
		Itens: []ItemSaidaInput{{ItemID: itemIDValido, Quantidade: 1}},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Saida.OrdemServicoID != ordemServicoIDValida || fake.recebido.OrdemServicoID != ordemServicoIDValida {
		t.Fatalf("resultado=%+v recebido=%+v", resultado, fake.recebido)
	}
}
