package seguranca

import (
	mecanicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	"golang.org/x/crypto/bcrypt"
)

// Bcrypt implementa os contratos de senha mantendo o algoritmo e o custo existentes.
type Bcrypt struct{}

func NewBcrypt() Bcrypt { return Bcrypt{} }

func (Bcrypt) Gerar(senha string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	return string(hash), err
}

func (Bcrypt) Verificar(hash, senha string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
}

var (
	_ mecanicoApplication.GeradorHashSenha  = Bcrypt{}
	_ segurancaApplication.VerificadorSenha = Bcrypt{}
)
