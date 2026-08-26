package mecanico_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mecanicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	mecanicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/mecanico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/mecanico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestAtualizarMecanico(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	const email = "mecanico.atualizar@oficina.local"
	const novoEmail = "mecanico.atualizado@oficina.local"
	const osID = "99000000-0000-0000-0000-000000000001"
	limparMecanico(ctx, t, db, email)
	limparMecanico(ctx, t, db, novoEmail)
	defer limparMecanico(ctx, t, db, email)
	defer limparMecanico(ctx, t, db, novoEmail)
	defer db.Exec(ctx, `DELETE FROM ordem_servico WHERE id = $1`, osID)

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"mecanicos:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	mecanicoRepository := mecanicoInfrastructure.NewPostgresRepository(db)
	cadastrar := segurancaPresentation.RequireScope(jwt, "mecanicos:escrever", presentation.NewCadastrarHandler(mecanicoApplication.NewCadastrar(mecanicoRepository)))
	atualizar := segurancaPresentation.RequireScope(jwt, "mecanicos:escrever", presentation.NewAtualizarHandler(mecanicoApplication.NewAtualizar(mecanicoRepository)))

	createRequest := httptest.NewRequest(http.MethodPost, "/mecanicos", strings.NewReader(`{"nome":"Mecânico Atualizar","email":"mecanico.atualizar@oficina.local","senha":"mecanico123456789","escopos":["clientes:ler"]}`))
	createRequest.Header.Set("Authorization", "Bearer "+token)
	createResponse := httptest.NewRecorder()
	cadastrar.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("cadastro status %d: %s", createResponse.Code, createResponse.Body.String())
	}
	var created struct{ ID string }
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `INSERT INTO ordem_servico (id, cliente_id, veiculo_id, mecanico_responsavel_id, placa_veiculo, status)
		VALUES ($1, '20000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', $2, 'ABC1D23', 'EM_EXECUCAO')`, osID, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	var version int
	var senhaHash string
	if err := db.QueryRow(ctx, `SELECT m.version, u.senha_hash FROM mecanico m JOIN usuario u ON u.id = m.usuario_id WHERE m.id = $1`, created.ID).Scan(&version, &senhaHash); err != nil {
		t.Fatal(err)
	}

	updateRequest := httptest.NewRequest(http.MethodPut, "/mecanicos/"+created.ID, strings.NewReader(`{"nome":"Mecânico Atualizado","email":"mecanico.atualizado@oficina.local","escopos":["os:ler","clientes:escrever"]}`))
	updateRequest.SetPathValue("mecanicoId", created.ID)
	updateRequest.Header.Set("Authorization", "Bearer "+token)
	updateRequest.Header.Set("If-Match", "1")
	updateResponse := httptest.NewRecorder()
	atualizar.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("atualizacao status %d: %s", updateResponse.Code, updateResponse.Body.String())
	}
	var got struct {
		Email   string
		Version int
		Escopos []string
	}
	if err := json.NewDecoder(updateResponse.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Email != novoEmail || got.Version != version+1 || len(got.Escopos) != 2 {
		t.Fatalf("resposta: %#v", got)
	}
	var persistedHash string
	var osMecanicoID string
	if err := db.QueryRow(ctx, `SELECT u.senha_hash FROM usuario u WHERE u.email = $1`, novoEmail).Scan(&persistedHash); err != nil {
		t.Fatal(err)
	}
	if persistedHash != senhaHash {
		t.Fatal("senha foi alterada")
	}
	if _, err := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt).Execute(ctx, novoEmail, "mecanico123456789"); err != nil {
		t.Fatalf("login apos atualizar falhou: %v", err)
	}
	if err := db.QueryRow(ctx, `SELECT mecanico_responsavel_id::text FROM ordem_servico WHERE id = $1`, osID).Scan(&osMecanicoID); err != nil || osMecanicoID != created.ID {
		t.Fatalf("vinculo OS: %q, erro: %v", osMecanicoID, err)
	}

	assertAtualizarStatus(t, atualizar, token, "x", "1", `{}`, http.StatusBadRequest)
	assertAtualizarStatus(t, atualizar, "", created.ID, "1", `{}`, http.StatusUnauthorized)
	semEscopo, _ := jwt.Gerar("usuario", []string{"clientes:ler"})
	assertAtualizarStatus(t, atualizar, semEscopo, created.ID, "1", `{}`, http.StatusForbidden)
	assertAtualizarStatus(t, atualizar, token, "22222222-2222-2222-2222-222222222222", "1", `{}`, http.StatusNotFound)
	assertAtualizarStatus(t, atualizar, token, created.ID, "2", `{"nome":"Mecânico","email":"mecanico@oficina.local","escopos":["os:ler"]}`, http.StatusConflict)
	assertAtualizarStatus(t, atualizar, token, created.ID, "1", `{"nome":"Mecânico","email":"mecanico.outro@oficina.local","escopos":["os:ler"]}`, http.StatusPreconditionFailed)
	assertAtualizarStatus(t, atualizar, token, created.ID, "", `{}`, http.StatusPreconditionRequired)
}

func assertAtualizarStatus(t *testing.T, handler http.Handler, token, id, ifMatch, body string, status int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPut, "/mecanicos/"+id, strings.NewReader(body))
	request.SetPathValue("mecanicoId", id)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if ifMatch != "" {
		request.Header.Set("If-Match", ifMatch)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != status {
		t.Fatalf("status esperado %d, obtido %d: %s", status, response.Code, response.Body.String())
	}
}
