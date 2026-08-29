package peca

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
)

type linhaPecaFake struct {
	fornecedorID *string
	fabricante   *string
	precoVenda   *string
	err          error
}

func (linha linhaPecaFake) Scan(destinos ...any) error {
	if linha.err != nil {
		return linha.err
	}
	*(destinos[0].(*string)) = "item-id"
	*(destinos[1].(*string)) = "PEC-000001"
	*(destinos[2].(*string)) = "Filtro de oleo"
	*(destinos[3].(*string)) = "Filtro original"
	*(destinos[4].(*string)) = "categoria-id"
	*(destinos[5].(*string)) = "Filtros"
	if linha.fornecedorID != nil {
		*(destinos[6].(*string)) = *linha.fornecedorID
	}
	*(destinos[7].(**string)) = linha.fabricante
	*(destinos[8].(*string)) = "UN"
	*(destinos[9].(**string)) = linha.precoVenda
	*(destinos[10].(*int64)) = 10
	*(destinos[11].(*int64)) = 2
	*(destinos[12].(*int64)) = 3
	*(destinos[13].(*bool)) = true
	*(destinos[14].(*int)) = 4
	*(destinos[15].(*bool)) = true
	return nil
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatorio")
	}
}

func TestLer(t *testing.T) {
	fornecedorID := "fornecedor-id"
	fabricante := "Bosch"
	preco := "99.90"
	item, err := ler(linhaPecaFake{fornecedorID: &fornecedorID, fabricante: &fabricante, precoVenda: &preco})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "item-id" || item.Codigo != "PEC-000001" || item.FornecedorID == nil || *item.FornecedorID != fornecedorID {
		t.Fatalf("peca invalida: %+v", item)
	}
	if item.Fabricante == nil || *item.Fabricante != fabricante || item.PrecoVenda == nil || *item.PrecoVenda != preco {
		t.Fatalf("campos opcionais invalidos: %+v", item)
	}

	item, err = ler(linhaPecaFake{})
	if err != nil || item.FornecedorID != nil {
		t.Fatalf("peca sem fornecedor invalida: %+v erro=%v", item, err)
	}

	if _, err = ler(linhaPecaFake{err: errors.New("db")}); err == nil {
		t.Fatal("esperava erro")
	}
}

func TestMontarCondicoes(t *testing.T) {
	condicoes, args := montarCondicoes(pecaApplication.Filtros{
		Codigo:             "PEC-000001",
		Descricao:          "oleo",
		CategoriaID:        "categoria-id",
		Fabricante:         "Bosch",
		SomenteDisponiveis: true,
	})
	if !strings.Contains(condicoes, "i.tipo = 'PECA'") || !strings.Contains(condicoes, "i.ativo") {
		t.Fatalf("condicoes base ausentes: %s", condicoes)
	}
	for _, trecho := range []string{"i.codigo = $1", "i.descricao ILIKE $2", "i.categoria_id = $3", "i.fabricante ILIKE $4", "(i.saldo_fisico - i.saldo_reservado) > 0"} {
		if !strings.Contains(condicoes, trecho) {
			t.Fatalf("condicao %q ausente em %s", trecho, condicoes)
		}
	}
	if len(args) != 4 || args[1] != "%oleo%" || args[3] != "%Bosch%" {
		t.Fatalf("args invalidos: %#v", args)
	}

	condicoes, _ = montarCondicoes(pecaApplication.Filtros{Codigo: "PEC-000001", IncluirInativos: true})
	if strings.Contains(condicoes, "i.ativo") {
		t.Fatalf("nao deveria filtrar ativos: %s", condicoes)
	}
}

func TestHelpers(t *testing.T) {
	if valorTexto(nil) != "0" {
		t.Fatal("valorTexto nil invalido")
	}
	preco := " 10.50 "
	if valorTexto(&preco) != "10.50" {
		t.Fatal("valorTexto deveria aparar espacos")
	}
	if valorParcial(&preco, 2) != 21 {
		t.Fatal("valorParcial invalido")
	}
	invalido := "abc"
	if valorParcial(&invalido, 2) != 0 {
		t.Fatal("valorParcial invalido deveria zerar")
	}
	if !mesmoPreco("10.00", "10") || mesmoPreco("10.01", "10") {
		t.Fatal("mesmoPreco invalido")
	}
	if !mesmoPreco("abc", "abc") || mesmoPreco("abc", "10") {
		t.Fatal("mesmoPreco texto invalido")
	}
	if valorFornecedor(nil) != nil {
		t.Fatal("valorFornecedor nil invalido")
	}
	if valorFornecedor(&preco) != preco {
		t.Fatal("valorFornecedor invalido")
	}
}

type consultorPecaFake struct{ row pgx.Row }

func (fake consultorPecaFake) QueryRow(context.Context, string, ...any) pgx.Row {
	return fake.row
}

type ativoRowFake struct {
	ativo bool
	err   error
}

func (row ativoRowFake) Scan(destinos ...any) error {
	if row.err != nil {
		return row.err
	}
	*(destinos[0].(*bool)) = row.ativo
	return nil
}

func TestValidarFornecedor(t *testing.T) {
	if err := validarFornecedor(context.Background(), consultorPecaFake{}, nil); err != nil {
		t.Fatalf("fornecedor opcional deveria ser aceito: %v", err)
	}

	fornecedorID := "fornecedor-id"
	if err := validarFornecedor(context.Background(), consultorPecaFake{row: ativoRowFake{ativo: true}}, &fornecedorID); err != nil {
		t.Fatalf("fornecedor ativo deveria ser aceito: %v", err)
	}
	if err := validarFornecedor(context.Background(), consultorPecaFake{row: ativoRowFake{ativo: false}}, &fornecedorID); !errors.Is(err, pecaApplication.ErrFornecedorInvalido) {
		t.Fatalf("erro=%v", err)
	}
	if err := validarFornecedor(context.Background(), consultorPecaFake{row: ativoRowFake{err: pgx.ErrNoRows}}, &fornecedorID); !errors.Is(err, pecaApplication.ErrFornecedorInvalido) {
		t.Fatalf("erro=%v", err)
	}
	if err := validarFornecedor(context.Background(), consultorPecaFake{row: ativoRowFake{err: errors.New("db")}}, &fornecedorID); err == nil || errors.Is(err, pecaApplication.ErrFornecedorInvalido) {
		t.Fatalf("erro=%v", err)
	}
}
