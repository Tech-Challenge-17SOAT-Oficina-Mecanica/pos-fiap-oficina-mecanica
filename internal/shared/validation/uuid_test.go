package validation

import "testing"

func TestIsUUID(t *testing.T) {
	if !IsUUID("20000000-0000-0000-0000-000000000001") {
		t.Fatal("UUID válido rejeitado")
	}
	if IsUUID("invalido") {
		t.Fatal("UUID inválido aceito")
	}
}
