package seguranca_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestAutenticarMecanico(t *testing.T) {
	db, err := database.Open(context.Background())
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	jwt, err := infrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	handler := presentation.NewLoginHandler(application.NewAutenticar(infrastructure.NewPostgresRepository(db), jwt))
	for _, test := range []struct {
		body   string
		status int
	}{
		{`{"email":"mecanico@oficina.local","senha":"mecanico123"}`, 200},
		{`{"email":"mecanico@oficina.local","senha":"invalida"}`, 401},
		{`{}`, 400},
	} {
		request := httptest.NewRequest(http.MethodPost, "/autenticacao/login", strings.NewReader(test.body))
		response := httptest.NewRecorder()
		handler(response, request)
		if response.Code != test.status {
			t.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
	}
}
