package seguranca

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWT(t *testing.T) {
	if _, err := NewJWT(""); err == nil {
		t.Fatal("esperava segredo obrigatório")
	}
	service, err := NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return time.Now() }
	raw, err := service.Gerar("usuario", []string{"veiculos:ler"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.Validar(raw)
	if err != nil || claims.UsuarioID != "usuario" || len(claims.Escopos) != 1 {
		t.Fatalf("claims: %#v, %v", claims, err)
	}
	clienteRaw, err := service.GerarCliente("cliente", "ordem-servico")
	if err != nil {
		t.Fatal(err)
	}
	clienteClaims, err := service.Validar(clienteRaw)
	if err != nil || clienteClaims.ClienteID != "cliente" || clienteClaims.OrdemServicoID != "ordem-servico" || len(clienteClaims.Escopos) != 2 || clienteClaims.Escopos[0] != "os:ler" || clienteClaims.Escopos[1] != "orcamentos:decidir" {
		t.Fatalf("claims cliente: %#v, %v", clienteClaims, err)
	}
	if _, err := service.Validar("invalido"); err == nil {
		t.Fatal("esperava token inválido")
	}
	other, _ := NewJWT("outro")
	if _, err := other.Validar(raw); err == nil {
		t.Fatal("esperava assinatura inválida")
	}
	wrongAlgorithm, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, jwtClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "usuario", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString([]byte("segredo"))
	if _, err := service.Validar(wrongAlgorithm); err == nil {
		t.Fatal("esperava algoritmo inválido")
	}
}
