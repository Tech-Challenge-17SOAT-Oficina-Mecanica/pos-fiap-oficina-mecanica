package orcamento

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatorio")
	}
}

func TestDecimalHelpers(t *testing.T) {
	if menorDecimal("2.500", "3") != "2.5" {
		t.Fatal("menorDecimal deveria retornar primeiro valor normalizado")
	}
	if menorDecimal("4", "3.250") != "3.25" {
		t.Fatal("menorDecimal deveria retornar segundo valor normalizado")
	}
	if subtrairDecimal("5.500", "2.125") != "3.375" {
		t.Fatal("subtrairDecimal invalido")
	}
	if compararDecimal("2.00", "2") != 0 || compararDecimal("1", "2") >= 0 || compararDecimal("3", "2") <= 0 {
		t.Fatal("compararDecimal invalido")
	}
	if decimal("abc").Cmp(new(big.Rat)) != 0 {
		t.Fatal("decimal invalido deveria zerar")
	}
	for entrada, esperado := range map[string]string{
		"3.500": "3.5",
		"3.000": "3",
		"0.000": "0",
		"-0":    "0",
	} {
		if normalizarDecimal(entrada) != esperado {
			t.Fatalf("normalizarDecimal(%q)=%q", entrada, normalizarDecimal(entrada))
		}
	}
}

func TestCustoUnitarioCompra(t *testing.T) {
	custo := "12.50"
	if custoUnitarioCompra(itemAprovacao{tipo: "INSUMO", custoUnitario: &custo}) != custo {
		t.Fatal("insumo deveria usar custo unitario")
	}
	if custoUnitarioCompra(itemAprovacao{tipo: "INSUMO"}) != nil {
		t.Fatal("insumo sem custo deveria retornar nil")
	}
	if custoUnitarioCompra(itemAprovacao{tipo: "PECA", custoUnitario: &custo}) != nil {
		t.Fatal("peca nao deveria usar custo unitario")
	}
}
