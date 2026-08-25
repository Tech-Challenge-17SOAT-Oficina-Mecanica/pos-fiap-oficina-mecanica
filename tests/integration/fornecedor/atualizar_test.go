package integration_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestAtualizarFornecedor(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	db, err := database.OpenPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	documento := cnpjValido(fmt.Sprintf("%012d", time.Now().UnixNano()%1000000000000))
	var fornecedorID string
	var version int
	err = db.QueryRow(ctx, `
		INSERT INTO fornecedor (razao_social, documento, tipo_documento, telefone, prazo_entrega_dias)
		VALUES ($1, $2, 'CNPJ', '11999990000', 7)
		RETURNING id, version`, "Fornecedor Atualizacao Ltda", documento).Scan(&fornecedorID, &version)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM fornecedor WHERE id = $1", fornecedorID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"compras:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	semPermissao, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"compras:ler"})
	if err != nil {
		t.Fatal(err)
	}

	repository := infrastructure.NewPostgresRepository(db)
	mux := http.NewServeMux()
	mux.Handle("PUT /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", presentation.NewAtualizarHandler(application.NewAtualizarFornecedor(repository))))

	tests := []struct {
		name    string
		body    string
		version string
		token   string
		status  int
		text    string
	}{
		{name: "atualiza", body: `{"razaoSocial":"Fornecedor Atualizado Ltda","telefone":"11888887777","prazoEntregaDias":10}`, version: fmt.Sprint(version), token: token, status: http.StatusOK, text: `"version":2`},
		{name: "if-match antigo", body: `{"razaoSocial":"Fornecedor Atualizado Ltda","telefone":"11888887777"}`, version: fmt.Sprint(version), token: token, status: http.StatusPreconditionFailed},
		{name: "sem if-match", body: `{}`, token: token, status: http.StatusPreconditionRequired},
		{name: "documento no corpo", body: `{"razaoSocial":"Fornecedor Atualizado Ltda","documento":"` + documento + `","telefone":"11888887777"}`, version: "2", token: token, status: http.StatusBadRequest},
		{name: "sem escopo", body: `{"razaoSocial":"Fornecedor Atualizado Ltda","telefone":"11888887777"}`, version: "2", token: semPermissao, status: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/fornecedores/"+fornecedorID, bytes.NewBufferString(test.body))
			request.SetPathValue("fornecedorId", fornecedorID)
			if test.version != "" {
				request.Header.Set("If-Match", test.version)
			}
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if test.text != "" && !strings.Contains(response.Body.String(), test.text) {
				t.Fatalf("body=%s", response.Body.String())
			}
		})
	}
}
