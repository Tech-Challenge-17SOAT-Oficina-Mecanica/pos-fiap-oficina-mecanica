package insumo

import (
	"math/big"
	"strings"
)

type Processamento struct {
	ItemID               string
	QuantidadeSolicitada string
	QuantidadeReservada  string
	QuantidadeCompra     string
	SaldoDisponivelApos  string
}

func NovoProcessamento(itemID, quantidade, saldoDisponivel string) Processamento {
	reservada := menorDecimal(quantidade, saldoDisponivel)
	if compararDecimal(reservada, "0") < 0 {
		reservada = "0"
	}
	return Processamento{
		ItemID:               itemID,
		QuantidadeSolicitada: quantidade,
		QuantidadeReservada:  reservada,
		QuantidadeCompra:     subtrairDecimal(quantidade, reservada),
		SaldoDisponivelApos:  subtrairDecimal(saldoDisponivel, reservada),
	}
}

func menorDecimal(a, b string) string {
	if compararDecimal(a, b) <= 0 {
		return normalizarDecimal(decimalProcessamento(a).FloatString(3))
	}
	return normalizarDecimal(decimalProcessamento(b).FloatString(3))
}

func subtrairDecimal(a, b string) string {
	resultado := decimalProcessamento(a)
	resultado.Sub(resultado, decimalProcessamento(b))
	return normalizarDecimal(resultado.FloatString(3))
}

func compararDecimal(a, b string) int {
	return decimalProcessamento(a).Cmp(decimalProcessamento(b))
}

func decimalProcessamento(valor string) *big.Rat {
	numero, ok := new(big.Rat).SetString(strings.TrimSpace(valor))
	if !ok {
		return new(big.Rat)
	}
	return numero
}
