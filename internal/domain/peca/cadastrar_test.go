package peca

import (
	"errors"
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

func TestNormalizarDescricaoAplicadaNoCadastro(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado string
	}{
		{"Pastilha de freio", "pastilha de freio"},
		{"ÓLEO Sintético", "oleo sintetico"},
		{"  correia   dentada  ", "correia dentada"},
		{"Amortecedor Dianteiro ESQUERDO", "amortecedor dianteiro esquerdo"},
		{"Junção Ação Coração", "juncao acao coracao"},
	}

	for _, caso := range casos {
		t.Run(caso.entrada, func(t *testing.T) {
			if obtido := validation.NormalizarDescricao(caso.entrada); obtido != caso.esperado {
				t.Fatalf("NormalizarDescricao(%q) = %q, esperado %q", caso.entrada, obtido, caso.esperado)
			}
		})
	}
}

func TestNovoCadastroValido(t *testing.T) {
	fabricante := "  Fabricante X  "
	preco := "180.00"
	estoque := int64(4)

	cadastro, err := NovoCadastro("Pastilha de freio", "Pastilha  de   freio dianteira",
		" 7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94 ", &fabricante, &preco, &estoque)
	if err != nil {
		t.Fatal(err)
	}

	if cadastro.Descricao != "Pastilha de freio dianteira" {
		t.Fatalf("descricao = %q; espaço duplo deveria ter sido colapsado", cadastro.Descricao)
	}
	if cadastro.DescricaoNormalizada != "pastilha de freio dianteira" {
		t.Fatalf("descricaoNormalizada = %q", cadastro.DescricaoNormalizada)
	}
	if cadastro.CategoriaID != "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94" {
		t.Fatalf("categoriaID = %q; deveria estar sem espaços", cadastro.CategoriaID)
	}
	if cadastro.Fabricante == nil || *cadastro.Fabricante != "Fabricante X" {
		t.Fatalf("fabricante = %v", cadastro.Fabricante)
	}
	if cadastro.UnidadeMedida != UnidadeMedidaPadrao {
		t.Fatalf("unidadeMedida = %q, esperado %q", cadastro.UnidadeMedida, UnidadeMedidaPadrao)
	}
	if cadastro.EstoqueMinimo != 4 {
		t.Fatalf("estoqueMinimo = %d", cadastro.EstoqueMinimo)
	}
}

func TestNovoCadastroAplicaPadroes(t *testing.T) {
	cadastro, err := NovoCadastro("Correia", "Correia dentada", "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94", nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cadastro.EstoqueMinimo != 0 {
		t.Fatalf("estoqueMinimo = %d, esperado 0 quando ausente", cadastro.EstoqueMinimo)
	}
	if cadastro.Fabricante != nil || cadastro.PrecoVenda != nil {
		t.Fatal("fabricante e precoVenda deveriam continuar nulos")
	}
}

func TestNovoCadastroInvalido(t *testing.T) {
	categoria := "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94"
	negativo := int64(-1)
	precoNegativo := "-1"
	precoNaoNumerico := "abc"
	fabricanteLongo := strings.Repeat("a", 151)

	casos := []struct {
		nome     string
		executar func() (Cadastro, error)
		esperado error
	}{
		{"nome vazio", func() (Cadastro, error) {
			return NovoCadastro("", "Descricao valida", categoria, nil, nil, nil)
		}, ErrNomeInvalido},
		{"nome longo", func() (Cadastro, error) {
			return NovoCadastro(strings.Repeat("a", 151), "Descricao valida", categoria, nil, nil, nil)
		}, ErrNomeInvalido},
		{"descricao curta", func() (Cadastro, error) {
			return NovoCadastro("Peca", "ab", categoria, nil, nil, nil)
		}, ErrDescricaoInvalida},
		{"categoria vazia", func() (Cadastro, error) {
			return NovoCadastro("Peca", "Descricao valida", "   ", nil, nil, nil)
		}, ErrCategoriaObrigatoria},
		{"fabricante longo", func() (Cadastro, error) {
			return NovoCadastro("Peca", "Descricao valida", categoria, &fabricanteLongo, nil, nil)
		}, ErrFabricanteInvalido},
		{"preco negativo", func() (Cadastro, error) {
			return NovoCadastro("Peca", "Descricao valida", categoria, nil, &precoNegativo, nil)
		}, ErrPrecoVendaInvalido},
		{"preco nao numerico", func() (Cadastro, error) {
			return NovoCadastro("Peca", "Descricao valida", categoria, nil, &precoNaoNumerico, nil)
		}, ErrPrecoVendaInvalido},
		{"estoque negativo", func() (Cadastro, error) {
			return NovoCadastro("Peca", "Descricao valida", categoria, nil, nil, &negativo)
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
