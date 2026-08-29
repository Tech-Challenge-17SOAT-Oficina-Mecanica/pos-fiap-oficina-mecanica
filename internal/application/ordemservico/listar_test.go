package ordemservico

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type listarRepositoryFake struct {
	itens    []domain.ItemListagem
	total    int
	err      error
	recebido domain.FiltrosListagem
}

func (fake *listarRepositoryFake) Listar(_ context.Context, filtros domain.FiltrosListagem, _, _ int) ([]domain.ItemListagem, int, error) {
	fake.recebido = filtros
	return fake.itens, fake.total, fake.err
}

func TestListarRejeitaStatusInvalido(t *testing.T) {
	useCase := NewListar(&listarRepositoryFake{})
	_, err := useCase.Execute(context.Background(), ListarInput{Status: "STATUS_INEXISTENTE"})
	if err != ErrStatusInvalido {
		t.Fatalf("erro=%v, esperado %v", err, ErrStatusInvalido)
	}
}

func TestListarRejeitaDocumentoInvalido(t *testing.T) {
	useCase := NewListar(&listarRepositoryFake{})
	_, err := useCase.Execute(context.Background(), ListarInput{Documento: "123"})
	if err != ErrDocumentoInvalido {
		t.Fatalf("erro=%v, esperado %v", err, ErrDocumentoInvalido)
	}
}

func TestListarRejeitaPlacaInvalida(t *testing.T) {
	useCase := NewListar(&listarRepositoryFake{})
	_, err := useCase.Execute(context.Background(), ListarInput{Placa: "###"})
	if err != ErrPlacaInvalida {
		t.Fatalf("erro=%v, esperado %v", err, ErrPlacaInvalida)
	}
}

func TestListarNormalizaFiltrosEDelegaAoRepositorio(t *testing.T) {
	fake := &listarRepositoryFake{itens: []domain.ItemListagem{{OrdemServicoID: "os-1"}}, total: 1}
	useCase := NewListar(fake)
	resultado, err := useCase.Execute(context.Background(), ListarInput{Status: "RECEBIDA", Documento: "390.533.447-05", Placa: "abc-1d23"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if fake.recebido.Documento != "39053344705" {
		t.Fatalf("documento nao normalizado: %q", fake.recebido.Documento)
	}
	if fake.recebido.Placa != "ABC1D23" {
		t.Fatalf("placa nao normalizada: %q", fake.recebido.Placa)
	}
	if resultado.TotalElementos != 1 || len(resultado.Itens) != 1 {
		t.Fatalf("resultado=%+v", resultado)
	}
}
