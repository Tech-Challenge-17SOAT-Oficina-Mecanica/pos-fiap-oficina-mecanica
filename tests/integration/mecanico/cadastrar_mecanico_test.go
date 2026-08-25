package mecanico_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	mecanicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	mecanicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/mecanico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/mecanico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestCadastrarMecanico(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	const email = "mecanico.integracao@oficina.local"
	limparMecanico(ctx, t, db, email)
	defer limparMecanico(ctx, t, db, email)

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"mecanicos:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	mecanicoRepository := mecanicoInfrastructure.NewPostgresRepository(db)
	handler := presentation.NewCadastrarHandler(mecanicoApplication.NewCadastrar(mecanicoRepository), jwt)

	request := httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(`{"nome":"Mecânico Integração","email":"mecanico.integracao@oficina.local","senha":"mecanico123456789","escopos":["clientes:ler"]}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	segurancaRepository := segurancaInfrastructure.NewPostgresRepository(db)
	if _, err := segurancaApplication.NewAutenticar(segurancaRepository, jwt).Execute(ctx, email, "mecanico123456789"); err != nil {
		t.Fatalf("novo login falhou: %v", err)
	}

	duplicado := httptest.NewRecorder()
	duplicateRequest := httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(`{"nome":"Mecânico Integração","email":"mecanico.integracao@oficina.local","senha":"mecanico123456789","escopos":["clientes:ler"]}`))
	duplicateRequest.Header.Set("Authorization", "Bearer "+token)
	handler(duplicado, duplicateRequest)
	if duplicado.Code != http.StatusConflict {
		t.Fatalf("duplicado status %d: %s", duplicado.Code, duplicado.Body.String())
	}

	semToken := httptest.NewRecorder()
	handler(semToken, httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(`{}`)))
	if semToken.Code != http.StatusUnauthorized {
		t.Fatalf("sem token status %d", semToken.Code)
	}

	semEscopoToken, err := jwt.Gerar("usuario", []string{"clientes:ler"})
	if err != nil {
		t.Fatal(err)
	}
	semEscopoRequest := httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(`{}`))
	semEscopoRequest.Header.Set("Authorization", "Bearer "+semEscopoToken)
	semEscopo := httptest.NewRecorder()
	handler(semEscopo, semEscopoRequest)
	if semEscopo.Code != http.StatusForbidden {
		t.Fatalf("sem escopo status %d", semEscopo.Code)
	}
}

func limparMecanico(ctx context.Context, t *testing.T, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, email string) {
	t.Helper()
	_, err := db.Exec(ctx, `DELETE FROM usuario_escopo WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(email) = lower($1))`, email)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `DELETE FROM mecanico WHERE usuario_id IN (SELECT id FROM usuario WHERE lower(email) = lower($1))`, email)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `DELETE FROM usuario WHERE lower(email) = lower($1)`, email)
	if err != nil {
		t.Fatal(err)
	}
}
