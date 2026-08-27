package servico_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	infra "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/servico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestDesativarEReativarServico(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	const (
		id          = "4f000000-0000-0000-0000-000000000011"
		duplicateID = "4f000000-0000-0000-0000-000000000012"
		usuarioID   = "90000000-0000-0000-0000-000000000001"
	)
	_, _ = db.Exec(ctx, `DELETE FROM servico WHERE id IN ($1, $2)`, id, duplicateID)
	defer db.Exec(ctx, `DELETE FROM servico WHERE id IN ($1, $2)`, id, duplicateID)
	_, err = db.Exec(ctx, `INSERT INTO servico
		(id, codigo, nome, nome_normalizado, descricao, valor, tempo_estimado_minutos, ativo, version, usuario_criacao)
		VALUES ($1, 'SER-999981', 'Serviço Situação Integração', 'servico situacao integracao', 'Teste', 80.00, 20, TRUE, 1, $2)`, id, usuarioID)
	if err != nil {
		t.Fatal(err)
	}

	jwt, err := seguranca.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.Gerar(usuarioID, []string{"servicos:escrever"})
	semEscopo, _ := jwt.Gerar(usuarioID, []string{"servicos:ler"})
	repository := infra.NewPostgresRepository(db)
	desativar := segurancaPresentation.RequireScope(jwt, "servicos:escrever", presentation.NewDesativarHandler(application.NewDesativar(repository)))
	reativar := segurancaPresentation.RequireScope(jwt, "servicos:escrever", presentation.NewReativarHandler(application.NewReativar(repository)))

	response := executarSituacao(desativar, http.MethodDelete, id, token)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ativo":false`) ||
		!strings.Contains(response.Body.String(), `"usuarioDesativacao":"`+usuarioID+`"`) {
		t.Fatalf("desativação %d: %s", response.Code, response.Body.String())
	}
	var quantidade int
	var ativo bool
	var usuarioDesativacao string
	if err := db.QueryRow(ctx, `SELECT COUNT(*), BOOL_AND(ativo), MAX(usuario_desativacao::text) FROM servico WHERE id = $1`, id).
		Scan(&quantidade, &ativo, &usuarioDesativacao); err != nil || quantidade != 1 || ativo || usuarioDesativacao != usuarioID {
		t.Fatalf("persistência: quantidade=%d ativo=%v usuario=%q erro=%v", quantidade, ativo, usuarioDesativacao, err)
	}
	if got := executarSituacao(desativar, http.MethodDelete, id, token); got.Code != http.StatusConflict {
		t.Fatalf("segunda desativação %d: %s", got.Code, got.Body.String())
	}
	_, err = db.Exec(ctx, `INSERT INTO servico
		(id, codigo, nome, nome_normalizado, valor, tempo_estimado_minutos, ativo, version, usuario_criacao)
		VALUES ($1, 'SER-999982', 'Serviço Situação Integração', 'servico situacao integracao', 90.00, 25, TRUE, 1, $2)`, duplicateID, usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if got := executarSituacao(reativar, http.MethodPost, id, token); got.Code != http.StatusConflict {
		t.Fatalf("reativação duplicada %d: %s", got.Code, got.Body.String())
	}
	if _, err := db.Exec(ctx, `DELETE FROM servico WHERE id = $1`, duplicateID); err != nil {
		t.Fatal(err)
	}

	response = executarSituacao(reativar, http.MethodPost, id, token)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ativo":true`) ||
		!strings.Contains(response.Body.String(), `"dataDesativacao":null`) {
		t.Fatalf("reativação %d: %s", response.Code, response.Body.String())
	}
	if got := executarSituacao(reativar, http.MethodPost, id, token); got.Code != http.StatusConflict {
		t.Fatalf("segunda reativação %d: %s", got.Code, got.Body.String())
	}
	if got := executarSituacao(desativar, http.MethodDelete, "abc", token); got.Code != http.StatusBadRequest {
		t.Fatalf("uuid inválido: %d", got.Code)
	}
	if got := executarSituacao(desativar, http.MethodDelete, id, semEscopo); got.Code != http.StatusForbidden {
		t.Fatalf("sem escopo: %d", got.Code)
	}
	if got := executarSituacao(desativar, http.MethodDelete, "4f000000-0000-0000-0000-0000000000ff", token); got.Code != http.StatusNotFound {
		t.Fatalf("não encontrado: %d", got.Code)
	}
}

func executarSituacao(handler http.Handler, method, id, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, "/servicos/"+id, nil)
	request.SetPathValue("servicoId", id)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
