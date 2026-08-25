package validation

import "testing"

func TestIsEmail(t *testing.T) {
	if !IsEmail("maria@oficina.local") || IsEmail("maria") {
		t.Fatal("validação de e-mail inválida")
	}
}
