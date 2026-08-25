package orcamento

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type repositoryFake struct {
	consulta domain.Consulta
	data     []domain.Consulta
	total    int
	err      error
	id       string
	doc      string
	offset   int
	limit    int
}

func (fake *repositoryFake) BuscarPorID(_ context.Context, id string) (domain.Consulta, error) {
	fake.id = id
	return fake.consulta, fake.err
}

func (fake *repositoryFake) BuscarPorDocumento(_ context.Context, documento string, offset, limit int) ([]domain.Consulta, int, error) {
	fake.doc, fake.offset, fake.limit = documento, offset, limit
	return fake.data, fake.total, fake.err
}

func TestConsultarPorID(t *testing.T) {
	repository := &repositoryFake{consulta: domain.Consulta{OrdemServicoID: "os"}}
	result, err := NewConsultar(repository).Execute(context.Background(), ConsultarInput{OrcamentoID: "orcamento", Pagina: 0, Tamanho: 20})
	if err != nil || result.Consulta == nil || result.Consulta.OrdemServicoID != "os" || repository.id != "orcamento" {
		t.Fatalf("resultado=%#v erro=%v", result, err)
	}
}

func TestConsultarPorDocumentoPaginado(t *testing.T) {
	repository := &repositoryFake{data: []domain.Consulta{{OrdemServicoID: "os"}}, total: 21}
	result, err := NewConsultar(repository).Execute(context.Background(), ConsultarInput{Documento: "39053344705", Pagina: 1, Tamanho: 20})
	if err != nil || result.TotalPaginas != 2 || repository.doc != "39053344705" || repository.offset != 20 || repository.limit != 20 {
		t.Fatalf("resultado=%#v repository=%#v erro=%v", result, repository, err)
	}
}

func TestConsultarValidacoes(t *testing.T) {
	cases := []struct {
		name  string
		input ConsultarInput
		want  error
	}{
		{"sem criterio", ConsultarInput{Pagina: 0, Tamanho: 20}, ErrCriterioObrigatorio},
		{"pagina negativa", ConsultarInput{OrcamentoID: "id", Pagina: -1, Tamanho: 20}, ErrPaginacaoInvalida},
		{"tamanho zero", ConsultarInput{OrcamentoID: "id", Pagina: 0, Tamanho: 0}, ErrPaginacaoInvalida},
		{"tamanho acima do limite", ConsultarInput{OrcamentoID: "id", Pagina: 0, Tamanho: 101}, ErrPaginacaoInvalida},
		{"documento invalido", ConsultarInput{Documento: "11111111111", Pagina: 0, Tamanho: 20}, errors.New("documento inválido")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewConsultar(&repositoryFake{}).Execute(context.Background(), test.input)
			if err == nil || err.Error() != test.want.Error() {
				t.Fatalf("erro=%v", err)
			}
		})
	}
}

func TestConsultarPropagaErro(t *testing.T) {
	want := errors.New("db")
	_, err := NewConsultar(&repositoryFake{err: want}).Execute(context.Background(), ConsultarInput{OrcamentoID: "id", Pagina: 0, Tamanho: 20})
	if !errors.Is(err, want) {
		t.Fatalf("erro=%v", err)
	}
}
