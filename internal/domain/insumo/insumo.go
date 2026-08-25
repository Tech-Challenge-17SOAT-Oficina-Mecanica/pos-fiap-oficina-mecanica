package insumo

import (
	"errors"
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
