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

func TestAtualizarCliente(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	const documentoAtual = "52998224725"
	const documentoNovo = "12345678909"
	const documentoDuplicado = "10000000019"
	const placa = "TST1A23"
	_, _ = db.Exec(ctx, `DELETE FROM veiculo WHERE placa = $1`, placa)
	_, _ = db.Exec(ctx, `DELETE FROM cliente WHERE documento IN ($1, $2, $3)`, documentoAtual, documentoNovo, documentoDuplicado)
	defer db.Exec(ctx, `DELETE FROM veiculo WHERE placa = $1`, placa)
	defer db.Exec(ctx, `DELETE FROM cliente WHERE documento IN ($1, $2, $3)`, documentoAtual, documentoNovo, documentoDuplicado)

	var clienteID string
	if err := db.QueryRow(ctx, `INSERT INTO cliente (nome, documento, tipo_documento, telefone, ativo, version) VALUES ('Cliente Atualizar', $1, 'CPF', '11988887777', TRUE, 1) RETURNING id`, documentoAtual).Scan(&clienteID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO cliente (nome, documento, tipo_documento, email, ativo, version) VALUES ('Cliente Duplicado', $1, 'CPF', 'dup@example.com', TRUE, 1)`, documentoDuplicado); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO veiculo (cliente_id, placa, marca, modelo, ano, ativo, version) VALUES ($1, $2, 'Fiat', 'Uno', 2015, TRUE, 1)`, clienteID, placa); err != nil {
		t.Fatal(err)
	}

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"clientes:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "clientes:escrever", presentation.NewAtualizarHandler(application.NewAtualizar(clienteInfrastructure.NewPostgresRepository(db))))
	body := `{"nome":"Cliente Atualizado","documento":"12345678909","tipoDocumento":"CPF","email":"cliente@example.com"}`

	request := putClienteRequest(clienteID, body, "Bearer "+token, "1")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"version":2`) {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var nome string
	var version, veiculos int
	if err := db.QueryRow(ctx, `SELECT nome, version FROM cliente WHERE id = $1`, clienteID).Scan(&nome, &version); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM veiculo WHERE cliente_id = $1 AND placa = $2`, clienteID, placa).Scan(&veiculos); err != nil {
		t.Fatal(err)
	}
	if nome != "Cliente Atualizado" || version != 2 || veiculos != 1 {
		t.Fatalf("nome=%q version=%d veiculos=%d", nome, version, veiculos)
	}

	for _, test := range []struct {
		name, id, body, auth, ifMatch string
		status                        int
	}{
		{"id ausente", "", body, "Bearer " + token, "2", http.StatusBadRequest},
		{"nome ausente", clienteID, `{"documento":"12345678909","tipoDocumento":"CPF","email":"cliente@example.com"}`, "Bearer " + token, "2", http.StatusBadRequest},
		{"documento ausente", clienteID, `{"nome":"Cliente","tipoDocumento":"CPF","email":"cliente@example.com"}`, "Bearer " + token, "2", http.StatusBadRequest},
		{"tipo ausente", clienteID, `{"nome":"Cliente","documento":"12345678909","email":"cliente@example.com"}`, "Bearer " + token, "2", http.StatusBadRequest},
		{"cpf invalido", clienteID, `{"nome":"Cliente","documento":"11111111111","tipoDocumento":"CPF","email":"cliente@example.com"}`, "Bearer " + token, "2", http.StatusBadRequest},
		{"sem contato", clienteID, `{"nome":"Cliente","documento":"12345678909","tipoDocumento":"CPF"}`, "Bearer " + token, "2", http.StatusBadRequest},
		{"nao encontrado", "00000000-0000-0000-0000-000000000000", body, "Bearer " + token, "2", http.StatusNotFound},
		{"duplicado", clienteID, `{"nome":"Cliente","documento":"10000000019","tipoDocumento":"CPF","email":"cliente@example.com"}`, "Bearer " + token, "2", http.StatusConflict},
		{"versao antiga", clienteID, body, "Bearer " + token, "1", http.StatusPreconditionFailed},
		{"sem if match", clienteID, body, "Bearer " + token, "", http.StatusPreconditionRequired},
		{"sem token", clienteID, body, "", "2", http.StatusUnauthorized},
		{"sem escopo", clienteID, body, "Bearer " + tokenClienteLer(t, jwt), "2", http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, putClienteRequest(test.id, test.body, test.auth, test.ifMatch))
			if response.Code != test.status {
				t.Fatalf("status %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func tokenClienteLer(t *testing.T, jwt segurancaInfrastructure.JWT) string {
	t.Helper()
	token, err := jwt.Gerar("usuario", []string{"clientes:ler"})
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func putClienteRequest(id, body, auth, ifMatch string) *http.Request {
	request := httptest.NewRequest(http.MethodPut, "/clientes/"+id, strings.NewReader(body))
	request.SetPathValue("clienteId", id)
	if auth != "" {
		request.Header.Set("Authorization", auth)
	}
	if ifMatch != "" {
		request.Header.Set("If-Match", ifMatch)
	}
	return request
}
