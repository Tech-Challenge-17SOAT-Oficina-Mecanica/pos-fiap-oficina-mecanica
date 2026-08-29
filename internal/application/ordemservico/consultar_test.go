package ordemservico

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type consultarRepositoryFake struct {
	resultado domain.ConsultaDetalhada
	err       error
	osID      string
	clienteID string
}

func (fake *consultarRepositoryFake) Consultar(_ context.Context, ordemServicoID, clienteID string) (domain.ConsultaDetalhada, error) {
	fake.osID, fake.clienteID = ordemServicoID, clienteID
	return fake.resultado, fake.err
}

func TestConsultarDelegaAoRepositorio(t *testing.T) {
	esperado := domain.ConsultaDetalhada{OrdemServicoID: "os-1", StatusOrdemServico: "RECEBIDA"}
	fake := &consultarRepositoryFake{resultado: esperado}
	useCase := NewConsultar(fake)

	resultado, err := useCase.Execute(context.Background(), "os-1", "cliente-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fake.osID != "os-1" || fake.clienteID != "cliente-1" {
		t.Fatalf("parametros repassados incorretamente: osID=%q clienteID=%q", fake.osID, fake.clienteID)
	}
	if resultado.StatusOrdemServico != "RECEBIDA" {
		t.Fatalf("status=%q", resultado.StatusOrdemServico)
	}
}

func TestConsultarPropagaErro(t *testing.T) {
	fake := &consultarRepositoryFake{err: ErrOrdemServicoNaoEncontrada}
	useCase := NewConsultar(fake)
	_, err := useCase.Execute(context.Background(), "os-1", "")
	if err != ErrOrdemServicoNaoEncontrada {
		t.Fatalf("erro=%v, esperado %v", err, ErrOrdemServicoNaoEncontrada)
	}
}
