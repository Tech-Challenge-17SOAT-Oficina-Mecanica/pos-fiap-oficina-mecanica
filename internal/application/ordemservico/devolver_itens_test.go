package ordemservico

import (
	"context"
	"testing"

	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

type devolverItensRepositoryFake struct {
	resultado domainEstoque.ResultadoDevolucao
	err       error
	osID      string
}

func (fake *devolverItensRepositoryFake) DevolverItensAoEstoque(_ context.Context, ordemServicoID string) (domainEstoque.ResultadoDevolucao, error) {
	fake.osID = ordemServicoID
	return fake.resultado, fake.err
}

func TestDevolverItensAoEstoqueDelegaAoRepositorio(t *testing.T) {
	esperado := domainEstoque.ResultadoDevolucao{
		OrdemServicoID:        "70000000-0000-0000-0000-000000000001",
		TotalItensProcessados: 2,
	}
	fake := &devolverItensRepositoryFake{resultado: esperado}
	useCase := NewDevolverItensAoEstoque(fake)

	resultado, err := useCase.Execute(context.Background(), esperado.OrdemServicoID)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fake.osID != esperado.OrdemServicoID {
		t.Fatalf("ordemServicoID=%q, esperado %q", fake.osID, esperado.OrdemServicoID)
	}
	if resultado.TotalItensProcessados != esperado.TotalItensProcessados {
		t.Fatalf("totalItensProcessados=%d, esperado %d", resultado.TotalItensProcessados, esperado.TotalItensProcessados)
	}
}
