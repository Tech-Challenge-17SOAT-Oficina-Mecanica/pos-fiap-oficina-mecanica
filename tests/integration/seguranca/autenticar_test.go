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
	"golang.org/x/crypto/bcrypt"
)

func TestAutenticarMecanico(t *testing.T) {
	db, err := database.Open(context.Background())
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(context.Background()); err != nil {
		t.Skip("banco indisponível")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("mecanico123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO usuario (id, email, senha_hash, ativo)
		VALUES ('90000000-0000-0000-0000-000000000001', 'mecanico@oficina.local', $1, TRUE)
		ON CONFLICT (email) DO UPDATE SET senha_hash = EXCLUDED.senha_hash, ativo = TRUE
	`, string(hash))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO mecanico (id, usuario_id, nome)
		VALUES ('90000000-0000-0000-0000-000000000002', '90000000-0000-0000-0000-000000000001', 'Mecânico Inicial')
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(context.Background(), `
		INSERT INTO usuario_escopo (usuario_id, escopo)
		VALUES ('90000000-0000-0000-0000-000000000001', 'mecanicos:escrever')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		t.Fatal(err)
	}
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
