package validation

import "testing"

func TestOnlyDigits(t *testing.T) {
	if got := OnlyDigits("04.252.011/0001-10"); got != "04252011000110" {
		t.Fatalf("documento=%q", got)
	}
}

func TestIsDocumento(t *testing.T) {
	for _, test := range []struct {
		documento, tipo string
		valid           bool
	}{
		{"39053344705", "cpf", true},
		{"04252011000110", "CNPJ", true},
		{"11111111111", "CPF", false},
		{"11111111111111", "CNPJ", false},
		{"123", "RG", false},
	} {
		if got := IsDocumento(test.documento, test.tipo); got != test.valid {
			t.Fatalf("IsDocumento(%q, %q)=%t", test.documento, test.tipo, got)
		}
	}
}

func TestMascararDocumento(t *testing.T) {
	for _, test := range []struct {
		documento, tipo, want string
	}{
		{"390.533.447-05", "CPF", "***.***.***-05"},
		{"04.252.011/0001-10", "CNPJ", "**.***.***/****-10"},
		{"1", "CPF", "*"},
		{"12", "outro", "**"},
	} {
		if got := MascararDocumento(test.documento, test.tipo); got != test.want {
			t.Fatalf("MascararDocumento(%q, %q)=%q", test.documento, test.tipo, got)
		}
	}
}
