package ordemservico

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type entregarRepositoryFake struct {
	resultado domain.ResultadoEntrega
	err       error
	recebido  EntregarInput
}

func (fake *entregarRepositoryFake) Entregar(_ context.Context, input EntregarInput) (domain.ResultadoEntrega, error) {
	fake.recebido = input
	return fake.resultado, fake.err
}

func TestEntregarNormalizaEntrada(t *testing.T) {
	fake := &entregarRepositoryFake{}
	useCase := NewEntregar(fake, nil, nil)
	_, err := useCase.Execute(context.Background(), EntregarInput{
		ClienteID: " 20000000-0000-0000-0000-000000000001 ", Observacoes: "  sem ressalvas  ",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fake.recebido.ClienteID != "20000000-0000-0000-0000-000000000001" || fake.recebido.Observacoes != "sem ressalvas" {
		t.Fatalf("entrada nao normalizada: %+v", fake.recebido)
	}
}

func TestEntregarRejeitaClienteIDInvalido(t *testing.T) {
	fake := &entregarRepositoryFake{}
	_, err := NewEntregar(fake, nil, nil).Execute(context.Background(), EntregarInput{ClienteID: "invalido"})
	if err != ErrClienteIDInvalido {
		t.Fatalf("erro=%v, esperado %v", err, ErrClienteIDInvalido)
	}
}

func TestEntregarDelegaAoRepositorio(t *testing.T) {
	esperado := domain.ResultadoEntrega{OrdemServicoID: "os-1", Status: domain.StatusEntregue}
	fake := &entregarRepositoryFake{resultado: esperado}
	resultado, err := NewEntregar(fake, nil, nil).Execute(context.Background(), EntregarInput{OSID: "os-1"})
	if err != nil || resultado.Status != domain.StatusEntregue {
		t.Fatalf("resultado=%+v erro=%v", resultado, err)
	}
}
