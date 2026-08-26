package seguranca

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
)

func TestClaimsAusentes(t *testing.T) {
	if claims := Claims(context.Background()); claims.UsuarioID != "" || len(claims.Escopos) != 0 {
		t.Fatalf("claims=%+v", claims)
	}
}

func TestRequireScope(t *testing.T) {
	jwt, err := infrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	allowed, _ := jwt.Gerar("usuario", []string{"veiculos:escrever"})
	denied, _ := jwt.Gerar("usuario", nil)
	for _, test := range []struct {
		header string
		status int
	}{{"", 401}, {"Bearer invalido", 401}, {"Bearer " + denied, 403}, {"Bearer " + allowed, 204}} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/", nil)
		request.Header.Set("Authorization", test.header)
		RequireScope(jwt, "veiculos:escrever", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if test.status == http.StatusNoContent && UsuarioID(request.Context()) != "usuario" {
				t.Fatal("usuário não foi propagado")
			}
			writer.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("header=%q status=%d", test.header, response.Code)
		}
	}
}
