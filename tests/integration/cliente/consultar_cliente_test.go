package cliente_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestConsultarCliente(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	const documentoSemVeiculo = "12345678909"
	_, _ = db.Exec(ctx, `DELETE FROM cliente WHERE documento = $1`, documentoSemVeiculo)
	_, err = db.Exec(ctx, `INSERT INTO cliente (nome, documento, tipo_documento, telefone, ativo, version) VALUES ('Cliente Sem Veiculo', $1, 'CPF', '11988887777', TRUE, 1)`, documentoSemVeiculo)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(ctx, `DELETE FROM cliente WHERE documento = $1`, documentoSemVeiculo)

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"clientes:ler"})
	if err != nil {
		t.Fatal(err)
	}
	handler := presentation.NewConsultarHandler(application.NewConsultar(clienteInfrastructure.NewPostgresRepository(db)), jwt)
	for _, test := range []struct {
		name, documento, auth, body string
		status                      int
	}{
		{"com veiculo", "39053344705", "Bearer " + token, `"veiculos":[{`, http.StatusOK},
		{"sem veiculo", documentoSemVeiculo, "Bearer " + token, `"veiculos":[]`, http.StatusOK},
		{"ausente", "", "Bearer " + token, "", http.StatusBadRequest},
		{"invalido", "11111111111", "Bearer " + token, "", http.StatusBadRequest},
		{"nao encontrado", "52998224725", "Bearer " + token, "", http.StatusNotFound},
		{"sem token", "39053344705", "", "", http.StatusUnauthorized},
		{"sem escopo", "39053344705", "Bearer " + tokenSemEscopo(t, jwt), "", http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/clientes?documento="+test.documento, nil)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			handler(response, request)
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
			if test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("body: %s", response.Body.String())
			}
		})
	}
}

func tokenSemEscopo(t *testing.T, jwt segurancaInfrastructure.JWT) string {
	t.Helper()
	token, err := jwt.Gerar("usuario", []string{"clientes:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	return token
}
