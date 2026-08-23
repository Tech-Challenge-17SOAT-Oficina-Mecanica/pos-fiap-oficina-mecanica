package veiculo

import "testing"

func TestNovoCadastro(t *testing.T) {
	for _, placa := range []string{"abc-1234", "abc 1d23"} {
		if _, err := NovoCadastro(placa, "Toyota", "Corolla", 2024); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := NovoCadastro("ABC123", "Toyota", "Corolla", 2024); err == nil {
		t.Fatal("placa inválida aceita")
	}
}
