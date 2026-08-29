package insumo

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	insumoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
)

type linhaInsumoFake struct {
	fornecedorID  *string
	custoUnitario *string
	err           error
}

func (linha linhaInsumoFake) Scan(destinos ...any) error {
	if linha.err != nil {
		return linha.err
	}
	*(destinos[0].(*string)) = "item-id"
	*(destinos[1].(*string)) = "INS-000001"
	*(destinos[2].(*string)) = "Oleo"
	*(destinos[3].(*string)) = "Oleo 5w30"
	*(destinos[4].(*string)) = "categoria-id"
	*(destinos[5].(*string)) = "Lubrificantes"
	if linha.fornecedorID != nil {
		*(destinos[6].(*string)) = *linha.fornecedorID
	}
	*(destinos[7].(*string)) = "L"
	*(destinos[8].(**string)) = linha.custoUnitario
	*(destinos[9].(*string)) = "10.500"
	*(destinos[10].(*string)) = "2.000"
	*(destinos[11].(*string)) = "3.000"
	*(destinos[12].(*bool)) = true
	*(destinos[13].(*int)) = 4
	*(destinos[14].(*bool)) = true
	return nil
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatorio")
	}
}

func TestLer(t *testing.T) {
	fornecedorID := "fornecedor-id"
	custo := "12.345"
	item, err := ler(linhaInsumoFake{fornecedorID: &fornecedorID, custoUnitario: &custo})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "item-id" || item.Codigo != "INS-000001" || item.FornecedorID == nil || *item.FornecedorID != fornecedorID {
		t.Fatalf("insumo invalido: %+v", item)
	}
	if item.CustoUnitario == nil || *item.CustoUnitario != custo || item.SaldoFisico != "10.500" {
		t.Fatalf("campos invalidos: %+v", item)
	}

	item, err = ler(linhaInsumoFake{})
	if err != nil || item.FornecedorID != nil {
		t.Fatalf("insumo sem fornecedor invalido: %+v erro=%v", item, err)
	}

	if _, err = ler(linhaInsumoFake{err: errors.New("db")}); err == nil {
		t.Fatal("esperava erro")
	}
}

func TestMontarCondicoes(t *testing.T) {
	quantidade := "2.500"
	condicoes, args := montarCondicoes(insumoApplication.FiltrosConsulta{
		Codigo:             "INS-000001",
		Descricao:          "oleo",
		CategoriaID:        "categoria-id",
		QuantidadeDesejada: &quantidade,
		SomenteDisponiveis: true,
	})
	if !strings.Contains(condicoes, "i.tipo = 'INSUMO'") || !strings.Contains(condicoes, "i.ativo") {
		t.Fatalf("condicoes base ausentes: %s", condicoes)
	}
	for _, trecho := range []string{"i.codigo = $1", "i.descricao ILIKE $2", "i.categoria_id = $3", "(i.saldo_fisico - i.saldo_reservado) >= $4::NUMERIC"} {
		if !strings.Contains(condicoes, trecho) {
			t.Fatalf("condicao %q ausente em %s", trecho, condicoes)
		}
	}
	if len(args) != 4 || args[1] != "%oleo%" || args[3] != quantidade {
		t.Fatalf("args invalidos: %#v", args)
	}

	condicoes, _ = montarCondicoes(insumoApplication.FiltrosConsulta{Codigo: "INS-000001", IncluirInativos: true})
	if strings.Contains(condicoes, "i.ativo") {
		t.Fatalf("nao deveria filtrar ativos: %s", condicoes)
	}
}

func TestHelpers(t *testing.T) {
	if valorTexto(nil) != "0" {
		t.Fatal("valorTexto nil invalido")
	}
	custo := " 4.50 "
	if valorTexto(&custo) != "4.50" {
		t.Fatal("valorTexto deveria aparar espacos")
	}
	if valorParcial(&custo, "2").String() != "9/1" {
		t.Fatal("valorParcial invalido")
	}
	if decimal("abc").Cmp(new(big.Rat)) != 0 {
		t.Fatal("decimal invalido deveria zerar")
	}
	if formatarValor(big.NewRat(9, 2)) != "4.50" {
		t.Fatal("formatarValor invalido")
	}
	if valorFornecedor(nil) != nil {
		t.Fatal("valorFornecedor nil invalido")
	}
	if valorFornecedor(&custo) != custo {
		t.Fatal("valorFornecedor invalido")
	}
}

type consultorInsumoFake struct{ row pgx.Row }

func (fake consultorInsumoFake) QueryRow(context.Context, string, ...any) pgx.Row {
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
	if err := validarFornecedor(context.Background(), consultorInsumoFake{}, nil); err != nil {
		t.Fatalf("fornecedor opcional deveria ser aceito: %v", err)
	}

	fornecedorID := "fornecedor-id"
	if err := validarFornecedor(context.Background(), consultorInsumoFake{row: ativoRowFake{ativo: true}}, &fornecedorID); err != nil {
		t.Fatalf("fornecedor ativo deveria ser aceito: %v", err)
	}
	if err := validarFornecedor(context.Background(), consultorInsumoFake{row: ativoRowFake{ativo: false}}, &fornecedorID); !errors.Is(err, insumoApplication.ErrFornecedorInvalido) {
		t.Fatalf("erro=%v", err)
	}
	if err := validarFornecedor(context.Background(), consultorInsumoFake{row: ativoRowFake{err: pgx.ErrNoRows}}, &fornecedorID); !errors.Is(err, insumoApplication.ErrFornecedorInvalido) {
		t.Fatalf("erro=%v", err)
	}
	if err := validarFornecedor(context.Background(), consultorInsumoFake{row: ativoRowFake{err: errors.New("db")}}, &fornecedorID); err == nil || errors.Is(err, insumoApplication.ErrFornecedorInvalido) {
		t.Fatalf("erro=%v", err)
	}
}
