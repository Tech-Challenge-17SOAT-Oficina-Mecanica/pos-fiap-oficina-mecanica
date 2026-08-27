package peca

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type repositorioFake struct {
	pecas          []peca.Peca
	total          int
	erro           error
	filtrosUsados  Filtros
	limiteUsado    int
	deslocaUsado   int
	chamouPorID    bool
	identificadorR string
}

func (fake *repositorioFake) BuscarPorFiltro(_ context.Context, filtros Filtros, limite, deslocamento int) ([]peca.Peca, int, error) {
	fake.filtrosUsados, fake.limiteUsado, fake.deslocaUsado = filtros, limite, deslocamento
	return fake.pecas, fake.total, fake.erro
}

func (fake *repositorioFake) BuscarPorID(_ context.Context, id string) (peca.Peca, error) {
	fake.chamouPorID, fake.identificadorR = true, id
	if len(fake.pecas) == 0 {
		return peca.Peca{}, ErrNaoEncontrada
	}
	return fake.pecas[0], fake.erro
}

func quantidade(valor int64) *int64 { return &valor }

func TestExecuteValidaFiltros(t *testing.T) {
	casos := []struct {
		nome    string
		filtros Filtros
		erro    error
	}{
		{"sem codigo nem descricao", Filtros{}, ErrFiltroObrigatorio},
		{"apenas espacos", Filtros{Codigo: "   "}, ErrFiltroObrigatorio},
		{"descricao curta", Filtros{Descricao: "a"}, ErrDescricaoCurta},
		{"categoria nao e uuid", Filtros{Codigo: "PEC-000001", CategoriaID: "abc"}, ErrIdentificadorInvalido},
		{"quantidade zero", Filtros{Codigo: "PEC-000001", QuantidadeDesejada: quantidade(0)}, ErrQuantidadeInvalida},
		{"quantidade negativa", Filtros{Codigo: "PEC-000001", QuantidadeDesejada: quantidade(-1)}, ErrQuantidadeInvalida},
		{"codigo valido", Filtros{Codigo: "PEC-000001"}, nil},
		{"descricao valida", Filtros{Descricao: "oleo"}, nil},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			useCase := NewConsultarPecas(&repositorioFake{})
			_, err := useCase.Execute(context.Background(), caso.filtros, 20, 0)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro = %v, esperado %v", err, caso.erro)
			}
		})
	}
}

func TestExecuteRepassaPaginacaoEFiltrosNormalizados(t *testing.T) {
	fake := &repositorioFake{pecas: []peca.Peca{{Codigo: "PEC-000001"}}, total: 1}
	useCase := NewConsultarPecas(fake)

	resultado, err := useCase.Execute(context.Background(), Filtros{Descricao: "  filtro  "}, 20, 40)
	if err != nil {
		t.Fatal(err)
	}
	if fake.filtrosUsados.Descricao != "filtro" {
		t.Fatalf("descricao = %q, esperada sem espacos", fake.filtrosUsados.Descricao)
	}
	if fake.limiteUsado != 20 || fake.deslocaUsado != 40 {
		t.Fatalf("limite/deslocamento = %d/%d, esperado 20/40", fake.limiteUsado, fake.deslocaUsado)
	}
	if resultado.TotalElementos != 1 || len(resultado.Pecas) != 1 {
		t.Fatalf("resultado inesperado: %+v", resultado)
	}
}

func TestExecuteSemResultadoNaoEErro(t *testing.T) {
	useCase := NewConsultarPecas(&repositorioFake{})

	resultado, err := useCase.Execute(context.Background(), Filtros{Codigo: "PEC-999999"}, 20, 0)
	if err != nil {
		t.Fatalf("lista vazia nao deve ser erro, recebido %v", err)
	}
	if resultado.TotalElementos != 0 || len(resultado.Pecas) != 0 {
		t.Fatalf("resultado deveria ser vazio: %+v", resultado)
	}
}

func TestBuscarPorIDValidaUUID(t *testing.T) {
	fake := &repositorioFake{}
	useCase := NewConsultarPecas(fake)

	if _, err := useCase.BuscarPorID(context.Background(), "nao-e-uuid"); !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v, esperado ErrIdentificadorInvalido", err)
	}
	if fake.chamouPorID {
		t.Fatal("repositorio nao deveria ser consultado com identificador invalido")
	}
}

func TestBuscarPorIDInexistente(t *testing.T) {
	useCase := NewConsultarPecas(&repositorioFake{})

	_, err := useCase.BuscarPorID(context.Background(), "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4")
	if !errors.Is(err, ErrNaoEncontrada) {
		t.Fatalf("erro = %v, esperado ErrNaoEncontrada", err)
	}
}
