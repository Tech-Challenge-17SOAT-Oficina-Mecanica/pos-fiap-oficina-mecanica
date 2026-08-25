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
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestCadastrarCliente(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	const documento = "52998224725"
	_, _ = db.Exec(ctx, `DELETE FROM cliente WHERE documento = $1`, documento)
	defer db.Exec(ctx, `DELETE FROM cliente WHERE documento = $1`, documento)

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"clientes:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "clientes:escrever", presentation.NewCadastrarHandler(application.NewCadastrar(clienteInfrastructure.NewPostgresRepository(db))))
	request := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(`{"nome":"Cliente Teste","documento":"52998224725","tipoDocumento":"CPF","telefone":"11988887777"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var nome string
	if err := db.QueryRow(ctx, `SELECT nome FROM cliente WHERE documento = $1 AND ativo`, documento).Scan(&nome); err != nil || nome != "Cliente Teste" {
		t.Fatalf("nome: %q, erro: %v", nome, err)
	}

	duplicateRequest := httptest.NewRequest(http.MethodPost, "/clientes", strings.NewReader(`{"nome":"Cliente Teste","documento":"52998224725","tipoDocumento":"CPF","telefone":"11988887777"}`))
	duplicateRequest.Header.Set("Authorization", "Bearer "+token)
	duplicado := httptest.NewRecorder()
	handler.ServeHTTP(duplicado, duplicateRequest)
	if duplicado.Code != http.StatusConflict {
		t.Fatalf("duplicado status %d", duplicado.Code)
	}
}
