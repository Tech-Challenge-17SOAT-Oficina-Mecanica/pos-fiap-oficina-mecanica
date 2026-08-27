package insumo

import (
	"errors"
	"testing"
)

func TestInsumoDesativar(t *testing.T) {
	desativado, err := (Insumo{ID: "insumo-1", Ativo: true}).Desativar()
	if err != nil || desativado.Ativo {
		t.Fatalf("insumo=%+v erro=%v", desativado, err)
	}
}

func TestInsumoDesativarJaInativo(t *testing.T) {
	_, err := (Insumo{ID: "insumo-1", Ativo: false}).Desativar()
	if !errors.Is(err, ErrInsumoJaInativo) {
		t.Fatalf("erro=%v", err)
	}
}
