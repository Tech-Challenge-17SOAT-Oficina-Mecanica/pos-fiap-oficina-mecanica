package fornecedor

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type consultaRepositoryStub struct {
	filtros FiltrosConsulta
	data    []domain.Fornecedor
	total   int
	err     error
}

func (stub *consultaRepositoryStub) Listar(_ context.Context, filtros FiltrosConsulta) ([]domain.Fornecedor, int, error) {
	stub.filtros = filtros
	return stub.data, stub.total, stub.err
}

func (stub *consultaRepositoryStub) BuscarPorID(_ context.Context, _ string) (domain.Fornecedor, error) {
	return domain.Fornecedor{}, nil
}

func TestConsultarFornecedoresNormalizaFiltrosEPagina(t *testing.T) {
	repository := &consultaRepositoryStub{total: 51}
	resultado, err := NewConsultarFornecedores(repository).Execute(context.Background(), FiltrosConsulta{
		Nome:      " Auto ",
		Documento: "04252011000110",
		Pagina:    1,
		Tamanho:   50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repository.filtros.Nome != "Auto" || repository.filtros.Documento != "04252011000110" {
		t.Fatalf("filtros=%+v", repository.filtros)
	}
	if resultado.TotalPaginas != 2 {
		t.Fatalf("totalPaginas=%d", resultado.TotalPaginas)
	}
}

func TestConsultarFornecedoresAplicaTamanhoPadrao(t *testing.T) {
	repository := &consultaRepositoryStub{}
	_, err := NewConsultarFornecedores(repository).Execute(context.Background(), FiltrosConsulta{})
	if err != nil {
		t.Fatal(err)
	}
	if repository.filtros.Tamanho != TamanhoPaginaPadrao {
		t.Fatalf("tamanho=%d", repository.filtros.Tamanho)
	}
}

func TestConsultarFornecedoresRejeitaPaginacaoInvalida(t *testing.T) {
	_, err := NewConsultarFornecedores(&consultaRepositoryStub{}).Execute(context.Background(), FiltrosConsulta{Tamanho: 51})
	if !errors.Is(err, ErrConsultaInvalida) {
		t.Fatalf("err=%v", err)
	}
}

func TestConsultarFornecedoresRejeitaDocumentoNaoNumerico(t *testing.T) {
	_, err := NewConsultarFornecedores(&consultaRepositoryStub{}).Execute(context.Background(), FiltrosConsulta{Documento: "04.252.011/0001-10"})
	if !errors.Is(err, ErrConsultaInvalida) {
		t.Fatalf("err=%v", err)
	}
}

func TestConsultarFornecedorPorIDRejeitaIdentificadorVazio(t *testing.T) {
	_, err := NewConsultarFornecedorPorID(&consultaRepositoryStub{}).Execute(context.Background(), " ")
	if !errors.Is(err, ErrConsultaInvalida) {
		t.Fatalf("err=%v", err)
	}
}
