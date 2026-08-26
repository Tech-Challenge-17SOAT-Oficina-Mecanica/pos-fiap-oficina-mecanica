package insumo

import (
	"errors"
	"strings"
	"testing"
)

const categoriaValida = "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724"

func ponteiro(valor string) *string { return &valor }

func TestNovoCadastroValido(t *testing.T) {
	cadastro, err := NovoCadastro("  Óleo 5W30  ", "Óleo   sintético  5W30 API SN",
		" "+categoriaValida+" ", " l ", ponteiro("45.0"), ponteiro("20.5"))
	if err != nil {
		t.Fatal(err)
	}

	if cadastro.Nome != "Óleo 5W30" {
		t.Fatalf("nome = %q", cadastro.Nome)
	}
	if cadastro.Descricao != "Óleo sintético 5W30 API SN" {
		t.Fatalf("descricao = %q; espaço duplo deveria ter sido colapsado", cadastro.Descricao)
	}
	if cadastro.DescricaoNormalizada != "oleo sintetico 5w30 api sn" {
		t.Fatalf("descricaoNormalizada = %q", cadastro.DescricaoNormalizada)
	}
	if cadastro.UnidadeMedida != "L" {
		t.Fatalf("unidadeMedida = %q; deveria ter sido normalizada para maiúscula", cadastro.UnidadeMedida)
	}
	if cadastro.CategoriaID != categoriaValida {
		t.Fatalf("categoriaID = %q", cadastro.CategoriaID)
	}
	if cadastro.EstoqueMinimo != "20.5" {
		t.Fatalf("estoqueMinimo = %q; a fração deveria ser preservada", cadastro.EstoqueMinimo)
	}
}

func TestNovoCadastroEstoqueMinimoPadrao(t *testing.T) {
	cadastro, err := NovoCadastro("Desengraxante", "Desengraxante multiuso", categoriaValida, "L", ponteiro("18"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if cadastro.EstoqueMinimo != "0" {
		t.Fatalf("estoqueMinimo = %q, esperado 0 quando ausente", cadastro.EstoqueMinimo)
	}
}

func TestNovoCadastroAceitaTodasAsUnidades(t *testing.T) {
	for _, unidade := range UnidadesMedida {
		t.Run(unidade, func(t *testing.T) {
			if _, err := NovoCadastro("Item", "Descricao valida", categoriaValida, unidade, ponteiro("1"), nil); err != nil {
				t.Fatalf("unidade %q deveria ser aceita: %v", unidade, err)
			}
		})
	}
}

func TestNovoCadastroInvalido(t *testing.T) {
	casos := []struct {
		nome     string
		executar func() (Cadastro, error)
		esperado error
	}{
		{"nome vazio", func() (Cadastro, error) {
			return NovoCadastro("", "Descricao valida", categoriaValida, "L", ponteiro("1"), nil)
		}, ErrNomeInvalido},
		{"nome longo", func() (Cadastro, error) {
			return NovoCadastro(strings.Repeat("a", 151), "Descricao valida", categoriaValida, "L", ponteiro("1"), nil)
		}, ErrNomeInvalido},
		{"descricao curta", func() (Cadastro, error) {
			return NovoCadastro("Item", "ab", categoriaValida, "L", ponteiro("1"), nil)
		}, ErrDescricaoInvalida},
		{"categoria vazia", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", "  ", "L", ponteiro("1"), nil)
		}, ErrCategoriaObrigatoria},
		{"unidade fora do enum", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", categoriaValida, "CX", ponteiro("1"), nil)
		}, ErrUnidadeInvalida},
		{"unidade vazia", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", categoriaValida, "", ponteiro("1"), nil)
		}, ErrUnidadeInvalida},
		{"custo ausente", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", categoriaValida, "L", nil, nil)
		}, ErrCustoInvalido},
		{"custo negativo", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", categoriaValida, "L", ponteiro("-1"), nil)
		}, ErrCustoInvalido},
		{"custo nao numerico", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", categoriaValida, "L", ponteiro("abc"), nil)
		}, ErrCustoInvalido},
		{"estoque negativo", func() (Cadastro, error) {
			return NovoCadastro("Item", "Descricao valida", categoriaValida, "L", ponteiro("1"), ponteiro("-0.5"))
		}, ErrEstoqueMinimoInvalido},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if _, err := caso.executar(); !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
		})
	}
}
