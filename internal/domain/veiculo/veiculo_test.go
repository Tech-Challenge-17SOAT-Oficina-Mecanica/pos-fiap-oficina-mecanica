package veiculo

import (
	"testing"
	"time"
)

func TestNovoCadastro(t *testing.T) {
	for _, placa := range []string{"abc-1234", "abc 1d23"} {
		if _, err := NovoCadastro(placa, "Toyota", "Corolla", 2024); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NovoCadastro("ABC123", "Toyota", "Corolla", 2024); err == nil {
		t.Fatal("placa inválida aceita")
	}
	for _, cadastro := range []struct {
		marca, modelo string
		ano           int
	}{
		{"", "Corolla", 2024},
		{"Toyota", "", 2024},
		{"Toyota", "Corolla", 1899},
		{"Toyota", "Corolla", time.Now().Year() + 2},
	} {
		if _, err := NovoCadastro("ABC1D23", cadastro.marca, cadastro.modelo, cadastro.ano); err == nil {
			t.Fatalf("cadastro inválido aceito: %+v", cadastro)
		}
	}
}

func TestNormalizarPlaca(t *testing.T) {
	if placa, err := NormalizarPlaca(" abc-1d23 "); err != nil || placa != "ABC1D23" {
		t.Fatalf("placa = %q, err = %v", placa, err)
	}
	if _, err := NormalizarPlaca(""); err == nil {
		t.Fatal("placa vazia aceita")
	}
}
