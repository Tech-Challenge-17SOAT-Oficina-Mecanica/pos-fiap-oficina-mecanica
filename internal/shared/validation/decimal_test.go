package validation

import "testing"

func TestDecimalPositivo(t *testing.T) {
	casos := map[string]bool{
		"1":      true,
		"1.234":  true,
		"0":      false,
		"0.000":  false,
		"-1":     false,
		"1.2345": false,
		"01":     false,
		"1,5":    false,
	}
	for valor, esperado := range casos {
		if got := DecimalPositivo(valor, 3); got != esperado {
			t.Fatalf("DecimalPositivo(%q) = %t, esperado %t", valor, got, esperado)
		}
	}
}

func TestDecimalNaoNegativo(t *testing.T) {
	if !DecimalNaoNegativo("0", 3) || !DecimalNaoNegativo("10.50", 3) {
		t.Fatal("zero e decimal positivo deveriam ser aceitos")
	}
	if DecimalNaoNegativo("-0.01", 3) || DecimalNaoNegativo("1.0001", 3) || DecimalNaoNegativo("abc", 3) {
		t.Fatal("decimal inválido foi aceito")
	}
}
