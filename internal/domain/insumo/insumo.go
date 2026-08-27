package insumo

import (
	"errors"
	"math/big"
	"strings"
	"time"
)

const Tipo = "INSUMO"

// Insumo admite quantidade fracionada — 10,0 L com consumo de 3,5 L resulta em 6,5 L —
// por isso saldos e estoque minimo sao decimais, diferente da peca, que usa inteiros.
// Os decimais trafegam como string para nao perder precisao no caminho ate o NUMERIC.
type Insumo struct {
	ID                   string
	Codigo               string
	Nome                 string
	Descricao            string
	CategoriaID          string
	Categoria            string
	UnidadeMedida        string
	CustoUnitario        *string
	SaldoFisico          string
	SaldoReservado       string
	EstoqueMinimo        string
	Ativo                bool
	DataDesativacao      *time.Time
	UsuarioDesativacao   *string
	Version              int
	PossuiPedidoEmAberto bool
	// DataCriacao so e carregada no cadastro.
	DataCriacao *time.Time
}

var ErrInsumoJaInativo = errors.New("insumo já está inativo")

func (item Insumo) Desativar() (Insumo, error) {
	if !item.Ativo {
		return Insumo{}, ErrInsumoJaInativo
	}
	item.Ativo = false
	return item, nil
}

func (insumo Insumo) SaldoDisponivel() string {
	return subtrair(insumo.SaldoFisico, insumo.SaldoReservado)
}

func (insumo Insumo) Disponivel() bool {
	return comparar(insumo.SaldoDisponivel(), "0") > 0
}

func (insumo Insumo) AbaixoDoMinimo() bool {
	return comparar(insumo.SaldoDisponivel(), insumo.EstoqueMinimo) < 0
}

func (insumo Insumo) AtendeQuantidade(quantidade string) bool {
	return comparar(insumo.SaldoDisponivel(), quantidade) >= 0
}

func subtrair(a, b string) string {
	decimalA := decimal(a)
	decimalA.Sub(decimalA, decimal(b))
	return normalizarDecimal(decimalA.FloatString(3))
}

func comparar(a, b string) int {
	return decimal(a).Cmp(decimal(b))
}

func decimal(valor string) *big.Rat {
	numero, ok := new(big.Rat).SetString(strings.TrimSpace(valor))
	if !ok {
		return new(big.Rat)
	}
	return numero
}

func normalizarDecimal(valor string) string {
	valor = strings.TrimRight(valor, "0")
	valor = strings.TrimRight(valor, ".")
	if valor == "" || valor == "-0" {
		return "0"
	}
	return valor
}
