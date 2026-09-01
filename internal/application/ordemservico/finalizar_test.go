package ordemservico

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type finalizarRepositoryFake struct {
	resultado domain.ResultadoFinalizacao
	err       error
	recebido  FinalizarInput
}

func (fake *finalizarRepositoryFake) Finalizar(_ context.Context, input FinalizarInput) (domain.ResultadoFinalizacao, error) {
	fake.recebido = input
	return fake.resultado, fake.err
}

func TestFinalizarNormalizaObservacoes(t *testing.T) {
	fake := &finalizarRepositoryFake{resultado: domain.ResultadoFinalizacao{OrdemServicoID: "os-1"}}
	useCase := NewFinalizar(fake, nil, nil)
	_, err := useCase.Execute(context.Background(), FinalizarInput{OSID: "os-1", Observacoes: "  tudo certo  "})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fake.recebido.Observacoes != "tudo certo" {
		t.Fatalf("observacoes=%q, esperado 'tudo certo'", fake.recebido.Observacoes)
	}
}

func TestFinalizarDelegaAoRepositorio(t *testing.T) {
	esperado := domain.ResultadoFinalizacao{OrdemServicoID: "os-1", Status: domain.StatusFinalizada}
	fake := &finalizarRepositoryFake{resultado: esperado}
	useCase := NewFinalizar(fake, nil, nil)
	resultado, err := useCase.Execute(context.Background(), FinalizarInput{OSID: "os-1"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Status != domain.StatusFinalizada {
		t.Fatalf("status=%q", resultado.Status)
	}
}

func TestFinalizarPropagaErro(t *testing.T) {
	fake := &finalizarRepositoryFake{err: domain.ErrOSNaoEmExecucao}
	useCase := NewFinalizar(fake, nil, nil)
	_, err := useCase.Execute(context.Background(), FinalizarInput{OSID: "os-1"})
	if err != domain.ErrOSNaoEmExecucao {
		t.Fatalf("erro=%v, esperado %v", err, domain.ErrOSNaoEmExecucao)
	}
}
