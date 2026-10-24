package seguranca

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptMantemAlgoritmoECusto(t *testing.T) {
	adapter := NewBcrypt()
	senha := "senha-de-teste-123"
	hash, err := adapter.Gerar(senha)
	if err != nil {
		t.Fatal(err)
	}
	if hash == senha {
		t.Fatal("senha nao foi transformada em hash")
	}
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil || cost != bcrypt.DefaultCost {
		t.Fatalf("custo=%d erro=%v", cost, err)
	}
	if err = adapter.Verificar(hash, senha); err != nil {
		t.Fatal(err)
	}
	if err = adapter.Verificar(hash, "senha-incorreta"); err == nil {
		t.Fatal("aceitou senha incorreta")
	}
	if err = adapter.Verificar("hash-invalido", senha); err == nil {
		t.Fatal("aceitou hash invalido")
	}
}

func TestBcryptRejeitaSenhaAcimaDoLimite(t *testing.T) {
	if _, err := NewBcrypt().Gerar(strings.Repeat("a", 73)); err == nil {
		t.Fatal("esperado erro de tamanho")
	}
}
