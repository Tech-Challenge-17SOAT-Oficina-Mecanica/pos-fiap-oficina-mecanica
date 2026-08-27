package peca

import (
	"errors"
	"testing"
	"time"
)

func TestDesativar(t *testing.T) {
	momento := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	original := Peca{ID: "1", Ativo: true, SaldoFisico: 9, SaldoReservado: 1}

	desativada, err := original.Desativar("usuario-1", momento)
	if err != nil {
		t.Fatal(err)
	}
	if desativada.Ativo {
		t.Fatal("peca deveria ficar inativa")
	}
	if !desativada.DataDesativacao.Equal(momento) || *desativada.UsuarioDesativacao != "usuario-1" {
		t.Fatalf("rastreabilidade incorreta: %+v", desativada)
	}
	if desativada.SaldoFisico != 9 || desativada.SaldoReservado != 1 {
		t.Fatalf("saldos nao podem mudar na desativacao: %+v", desativada)
	}
	if !original.Ativo {
		t.Fatal("Desativar nao deve alterar a peca original")
	}
}

func TestDesativarPecaJaInativa(t *testing.T) {
	_, err := Peca{ID: "1", Ativo: false}.Desativar("usuario-1", time.Now())
	if !errors.Is(err, ErrJaInativa) {
		t.Fatalf("erro = %v, esperado ErrJaInativa", err)
	}
}
