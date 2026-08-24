package cliente_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestDeletarCliente(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	const documento = "10000000019"
	const documentoReativar = "10000000108"
	const documentoComOS = "10000000289"
	const placa = "DEL1A23"
	const placaReativar = "REA1A23"
	const placaComOS = "OSA1A23"
	limparClienteTeste(ctx, db, []string{documento, documentoReativar, documentoComOS}, []string{placa, placaReativar, placaComOS})
	defer limparClienteTeste(ctx, db, []string{documento, documentoReativar, documentoComOS}, []string{placa, placaReativar, placaComOS})

	clienteID, veiculoID := inserirClienteVeiculo(t, ctx, db, documento, placa)
	reativarID, _ := inserirClienteVeiculo(t, ctx, db, documentoReativar, placaReativar)
	clienteComOSID, veiculoComOSID := inserirClienteVeiculo(t, ctx, db, documentoComOS, placaComOS)
	if _, err := db.Exec(ctx, `INSERT INTO ordem_servico (cliente_id, veiculo_id, placa_veiculo, status) VALUES ($1, $2, $3, 'EM_EXECUCAO')`, clienteComOSID, veiculoComOSID, placaComOS); err != nil {
		t.Fatal(err)
	}

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("00000000-0000-0000-0000-000000000001", []string{"clientes:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	repository := clienteInfrastructure.NewPostgresRepository(db)
	inativarHandler := presentation.NewInativarHandler(application.NewInativar(repository), jwt)
	reativarHandler := presentation.NewReativarHandler(application.NewReativar(repository), jwt)

	response := httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest(clienteID, "Bearer "+token, "duplicado"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"veiculosInativados":[{`) {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var clienteAtivo, veiculoAtivo bool
	var motivo string
	if err := db.QueryRow(ctx, `SELECT ativo, COALESCE(motivo_inativacao, '') FROM cliente WHERE id = $1`, clienteID).Scan(&clienteAtivo, &motivo); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT ativo FROM veiculo WHERE id = $1`, veiculoID).Scan(&veiculoAtivo); err != nil {
		t.Fatal(err)
	}
	if clienteAtivo || veiculoAtivo || motivo != "duplicado" {
		t.Fatalf("clienteAtivo=%v veiculoAtivo=%v motivo=%q", clienteAtivo, veiculoAtivo, motivo)
	}
	if _, err := db.Exec(ctx, `INSERT INTO cliente (nome, documento, tipo_documento, telefone, ativo, version) VALUES ('Cliente Recadastrado', $1, 'CPF', '11988887777', TRUE, 1)`, documento); err != nil {
		t.Fatalf("recadastro falhou: %v", err)
	}

	response = httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest(clienteID, "Bearer "+token, "duplicado"))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status idempotente %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest(clienteComOSID, "Bearer "+token, ""))
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"status":"EM_EXECUCAO"`) {
		t.Fatalf("status os aberta %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest("00000000-0000-0000-0000-000000000000", "Bearer "+token, ""))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status inexistente %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest(clienteComOSID, "Bearer "+tokenClienteLer(t, jwt), ""))
	if response.Code != http.StatusForbidden {
		t.Fatalf("status sem escopo %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest("id", "Bearer "+token, ""))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status uuid %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	inativarHandler(response, deleteClienteRequest(reativarID, "Bearer "+token, ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status inativar reativacao %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	reativarHandler(response, postReativacaoRequest(reativarID, "Bearer "+token))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"veiculosReativados":0`) {
		t.Fatalf("status reativar %d: %s", response.Code, response.Body.String())
	}
	if err := db.QueryRow(ctx, `SELECT ativo FROM veiculo WHERE placa = $1`, placaReativar).Scan(&veiculoAtivo); err != nil {
		t.Fatal(err)
	}
	if veiculoAtivo {
		t.Fatal("veículo foi reativado em cascata")
	}

	response = httptest.NewRecorder()
	reativarHandler(response, postReativacaoRequest(clienteID, "Bearer "+token))
	if response.Code != http.StatusConflict {
		t.Fatalf("status reativar duplicado %d: %s", response.Code, response.Body.String())
	}
}

func inserirClienteVeiculo(t *testing.T, ctx context.Context, db *pgxpool.Pool, documento, placa string) (string, string) {
	t.Helper()
	var clienteID string
	if err := db.QueryRow(ctx, `INSERT INTO cliente (nome, documento, tipo_documento, telefone, ativo, version) VALUES ('Cliente Teste', $1, 'CPF', '11988887777', TRUE, 1) RETURNING id`, documento).Scan(&clienteID); err != nil {
		t.Fatal(err)
	}
	var veiculoID string
	if err := db.QueryRow(ctx, `INSERT INTO veiculo (cliente_id, placa, marca, modelo, ano, ativo, version) VALUES ($1, $2, 'Fiat', 'Uno', 2015, TRUE, 1) RETURNING id`, clienteID, placa).Scan(&veiculoID); err != nil {
		t.Fatal(err)
	}
	return clienteID, veiculoID
}

func limparClienteTeste(ctx context.Context, db *pgxpool.Pool, documentos, placas []string) {
	_, _ = db.Exec(ctx, `DELETE FROM ordem_servico WHERE placa_veiculo IN ($1, $2, $3)`, placas[0], placas[1], placas[2])
	_, _ = db.Exec(ctx, `DELETE FROM veiculo WHERE placa IN ($1, $2, $3)`, placas[0], placas[1], placas[2])
	_, _ = db.Exec(ctx, `DELETE FROM cliente WHERE documento IN ($1, $2, $3)`, documentos[0], documentos[1], documentos[2])
}

func deleteClienteRequest(id, auth, motivo string) *http.Request {
	request := httptest.NewRequest(http.MethodDelete, "/clientes/"+id+"?motivo="+motivo, nil)
	request.SetPathValue("clienteId", id)
	if auth != "" {
		request.Header.Set("Authorization", auth)
	}
	return request
}

func postReativacaoRequest(id, auth string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/clientes/"+id+"/reativacao", nil)
	request.SetPathValue("clienteId", id)
	if auth != "" {
		request.Header.Set("Authorization", auth)
	}
	return request
}
