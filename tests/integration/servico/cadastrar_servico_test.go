package servico_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	infra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/servico"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestCadastrarServico(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	const nomeNormalizado = "servico de integracao"
	_, _ = db.Exec(ctx, `DELETE FROM servico WHERE nome_normalizado = $1`, nomeNormalizado)
	defer db.Exec(ctx, `DELETE FROM servico WHERE nome_normalizado = $1`, nomeNormalizado)

	jwt, err := seguranca.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"servicos:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := presentation.NewCadastrarHandler(application.NewCadastrar(infra.NewPostgresRepository(db)), jwt)

	request := httptest.NewRequest(http.MethodPost, "/servicos", strings.NewReader(
		`{"nome":"Serviço de Integração","descricao":"Teste completo","valor":199.90,"tempoEstimadoMinutos":45}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var output struct {
		ID, Codigo, Nome string
		Ativo            bool
		Version          int
	}
	if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
		t.Fatal(err)
	}
	if output.ID == "" || !strings.HasPrefix(output.Codigo, "SER-") || !output.Ativo || output.Version != 1 {
		t.Fatalf("resposta inesperada: %+v", output)
	}

	var usuarioCriacao string
	if err := db.QueryRow(ctx, `SELECT usuario_criacao::text FROM servico WHERE id = $1`, output.ID).Scan(&usuarioCriacao); err != nil || usuarioCriacao == "" {
		t.Fatalf("usuário de criação: %q, erro: %v", usuarioCriacao, err)
	}

	duplicate := httptest.NewRequest(http.MethodPost, "/servicos", strings.NewReader(
		`{"nome":"  SERVICO  DE INTEGRAÇÃO ","valor":10,"tempoEstimadoMinutos":5}`))
	duplicate.Header.Set("Authorization", "Bearer "+token)
	duplicateResponse := httptest.NewRecorder()
	handler(duplicateResponse, duplicate)
	if duplicateResponse.Code != http.StatusConflict {
		t.Fatalf("duplicado status %d: %s", duplicateResponse.Code, duplicateResponse.Body.String())
	}
}
