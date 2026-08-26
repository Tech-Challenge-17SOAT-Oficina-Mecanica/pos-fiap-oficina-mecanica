package servico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type consultarRepositoryFake struct {
	servicos []domain.Servico
	total    int
	servico  domain.Servico
	err      error
	filtros  Filtros
}

func (fake *consultarRepositoryFake) Listar(_ context.Context, filtros Filtros) ([]domain.Servico, int, error) {
	fake.filtros = filtros
	return fake.servicos, fake.total, fake.err
}

func (fake *consultarRepositoryFake) BuscarPorID(context.Context, string) (domain.Servico, error) {
	return fake.servico, fake.err
}

func TestConsultarListar(t *testing.T) {
	t.Run("lista paginada e normaliza filtro", func(t *testing.T) {
		repository := &consultarRepositoryFake{servicos: []domain.Servico{{ID: "id"}}, total: 21}
		pagina, err := NewConsultar(repository).Listar(context.Background(), Filtros{Nome: "  óleo ", Pagina: 1, Tamanho: 20})
		if err != nil || pagina.TotalPaginas != 2 || pagina.TotalElementos != 21 || repository.filtros.Nome != "óleo" {
			t.Fatalf("página: %+v, filtros: %+v, erro: %v", pagina, repository.filtros, err)
		}
	})

	t.Run("lista vazia não retorna null", func(t *testing.T) {
		pagina, err := NewConsultar(&consultarRepositoryFake{}).Listar(context.Background(), Filtros{Tamanho: 20})
		if err != nil || pagina.Servicos == nil || pagina.TotalPaginas != 0 {
			t.Fatalf("página: %+v, erro: %v", pagina, err)
		}
	})

	cases := []struct {
		name    string
		filtros Filtros
		err     error
	}{
		{"página negativa", Filtros{Pagina: -1, Tamanho: 20}, ErrPaginaInvalida},
		{"tamanho zero", Filtros{}, ErrTamanhoInvalido},
		{"tamanho acima do teto", Filtros{Tamanho: 51}, ErrTamanhoInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewConsultar(&consultarRepositoryFake{}).Listar(context.Background(), test.filtros)
			if !errors.Is(err, test.err) {
				t.Fatalf("erro: %v", err)
			}
		})
	}

	t.Run("erro do repositório", func(t *testing.T) {
		expected := errors.New("db")
		_, err := NewConsultar(&consultarRepositoryFake{err: expected}).Listar(context.Background(), Filtros{Tamanho: 20})
		if !errors.Is(err, expected) {
			t.Fatalf("erro: %v", err)
		}
	})
}

func TestConsultarPorID(t *testing.T) {
	expected := domain.Servico{ID: "id", Version: 2}
	got, err := NewConsultar(&consultarRepositoryFake{servico: expected}).PorID(context.Background(), "id")
	if err != nil || got != expected {
		t.Fatalf("serviço: %+v, erro: %v", got, err)
	}
}
