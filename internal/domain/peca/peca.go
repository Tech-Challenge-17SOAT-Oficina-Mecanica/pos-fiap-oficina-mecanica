package peca

import (
	"errors"
	"time"
)

var ErrJaInativa = errors.New("peca ja esta desativada")

type Peca struct {
	ID                   string
	Codigo               string
	Nome                 string
	Descricao            string
	CategoriaID          string
	Categoria            string
	Fabricante           *string
	UnidadeMedida        string
	PrecoVenda           *string
	SaldoFisico          int64
	SaldoReservado       int64
	EstoqueMinimo        int64
	Ativo                bool
	DataDesativacao      *time.Time
	UsuarioDesativacao   *string
	Version              int
	PossuiPedidoEmAberto bool
}

func (peca Peca) Desativar(usuarioID string, momento time.Time) (Peca, error) {
	if !peca.Ativo {
		return Peca{}, ErrJaInativa
	}
	peca.Ativo = false
	peca.DataDesativacao = &momento
	peca.UsuarioDesativacao = &usuarioID
	return peca, nil
}

func (peca Peca) SaldoDisponivel() int64 {
	return peca.SaldoFisico - peca.SaldoReservado
}

func (peca Peca) Disponivel() bool {
	return peca.SaldoDisponivel() > 0
}

func (peca Peca) AbaixoDoMinimo() bool {
	return peca.SaldoDisponivel() < peca.EstoqueMinimo
}

func (peca Peca) AtendeQuantidade(quantidade int64) bool {
	return peca.SaldoDisponivel() >= quantidade
}
