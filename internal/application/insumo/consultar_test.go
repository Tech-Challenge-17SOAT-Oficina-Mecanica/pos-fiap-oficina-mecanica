package insumo

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

type consultarRepositoryFake struct {
	insumos       []insumo.Insumo
	total         int
	erro          error
	filtrosUsados FiltrosConsulta
	limiteUsado   int
	offsetUsado   int
	chamouPorID   bool
}

func (fake *consultarRepositoryFake) BuscarPorFiltro(_ context.Context, filtros FiltrosConsulta, limite, deslocamento int) ([]insumo.Insumo, int, error) {
	fake.filtrosUsados, fake.limiteUsado, fake.offsetUsado = filtros, limite, deslocamento
	return fake.insumos, fake.total, fake.erro
}

func (fake *consultarRepositoryFake) BuscarPorID(context.Context, string) (insumo.Insumo, error) {
	fake.chamouPorID = true
	if len(fake.insumos) == 0 {
		return insumo.Insumo{}, ErrInsumoNaoEncontrado
	}
	return fake.insumos[0], fake.erro
}

func texto(valor string) *string { return &valor }

func TestConsultarInsumosValidaFiltros(t *testing.T) {
	casos := []struct {
		nome    string
		filtros FiltrosConsulta
		erro    error
	}{
		{"sem filtro", FiltrosConsulta{}, ErrFiltroObrigatorio},
		{"descricao curta", FiltrosConsulta{Descricao: "o"}, ErrDescricaoCurta},
		{"categoria invalida", FiltrosConsulta{CategoriaID: "abc"}, ErrIdentificadorInvalido},
		{"quantidade zero", FiltrosConsulta{Codigo: "INS-000001", QuantidadeDesejada: texto("0")}, ErrQuantidadeInvalida},
		{"quantidade com precisao invalida", FiltrosConsulta{Codigo: "INS-000001", QuantidadeDesejada: texto("1.0001")}, ErrQuantidadeInvalida},
		{"somente disponiveis sem quantidade", FiltrosConsulta{Codigo: "INS-000001", SomenteDisponiveis: true}, ErrQuantidadeObrigatoria},
		{"codigo valido", FiltrosConsulta{Codigo: " INS-000001 "}, nil},
		{"categoria valida", FiltrosConsulta{CategoriaID: categoriaValida}, nil},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			_, err := NewConsultarInsumos(&consultarRepositoryFake{}).Execute(context.Background(), caso.filtros, 20, 0)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro = %v, esperado %v", err, caso.erro)
			}
		})
	}
}

func TestConsultarInsumosRepassaFiltrosNormalizadosEPaginacao(t *testing.T) {
	fake := &consultarRepositoryFake{insumos: []insumo.Insumo{{Codigo: "INS-000001"}}, total: 1}
	resultado, err := NewConsultarInsumos(fake).Execute(context.Background(), FiltrosConsulta{Descricao: "  oleo  "}, 10, 20)
	if err != nil {
		t.Fatal(err)
	}
	if fake.filtrosUsados.Descricao != "oleo" || fake.limiteUsado != 10 || fake.offsetUsado != 20 {
		t.Fatalf("filtros/paginacao = %+v %d/%d", fake.filtrosUsados, fake.limiteUsado, fake.offsetUsado)
	}
	if resultado.TotalElementos != 1 || len(resultado.Insumos) != 1 {
		t.Fatalf("resultado = %+v", resultado)
	}
}

func TestBuscarInsumoPorIDValidaUUID(t *testing.T) {
	fake := &consultarRepositoryFake{}
	_, err := NewConsultarInsumos(fake).BuscarPorID(context.Background(), "abc")
	if !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v", err)
	}
	if fake.chamouPorID {
		t.Fatal("repositorio não deveria ser chamado com UUID inválido")
	}
}
