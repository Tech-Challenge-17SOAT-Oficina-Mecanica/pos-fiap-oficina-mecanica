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
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestAtualizarServico(t *testing.T) {
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
		id          = "4f000000-0000-0000-0000-000000000001"
		duplicateID = "4f000000-0000-0000-0000-000000000002"
		usuarioID   = "90000000-0000-0000-0000-000000000001"
	)
	_, _ = db.Exec(ctx, `DELETE FROM servico WHERE id IN ($1, $2)`, id, duplicateID)
	defer db.Exec(ctx, `DELETE FROM servico WHERE id IN ($1, $2)`, id, duplicateID)
	_, err = db.Exec(ctx, `
		INSERT INTO servico (id, codigo, nome, nome_normalizado, descricao, valor, tempo_estimado_minutos, ativo, version, usuario_criacao)
		VALUES ($1, 'SER-999991', 'Serviço Atualização Integração', 'servico atualizacao integracao', 'Original', 100.00, 30, TRUE, 1, $3),
		       ($2, 'SER-999992', 'Serviço Duplicado Integração', 'servico duplicado integracao', 'Duplicado', 50.00, 20, TRUE, 1, $3)`,
		id, duplicateID, usuarioID)
	if err != nil {
		t.Fatal(err)
	}

	jwt, err := seguranca.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.Gerar(usuarioID, []string{"servicos:escrever"})
	semEscopo, _ := jwt.Gerar(usuarioID, []string{"servicos:ler"})
	handler := segurancaPresentation.RequireScope(jwt, "servicos:escrever", presentation.NewAtualizarHandler(application.NewAtualizar(infra.NewPostgresRepository(db))))

	response := executarPATCH(handler, id, `{"descricao":"Descrição atualizada"}`, token, "1")
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var output struct {
		ID, Codigo, Nome, Descricao, DataAtualizacao string
		Valor                                        json.Number
		TempoEstimadoMinutos, Version                int
		Ativo                                        bool
	}
	if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
		t.Fatal(err)
	}
	if output.ID != id || output.Codigo != "SER-999991" || output.Nome != "Serviço Atualização Integração" ||
		output.Descricao != "Descrição atualizada" || output.Valor.String() != "100.00" ||
		output.TempoEstimadoMinutos != 30 || !output.Ativo || output.Version != 2 || output.DataAtualizacao == "" {
		t.Fatalf("resposta inesperada: %+v", output)
	}
	var version int
	var usuarioAtualizacao string
	if err := db.QueryRow(ctx, `SELECT version, usuario_atualizacao::text FROM servico WHERE id = $1`, id).
		Scan(&version, &usuarioAtualizacao); err != nil || version != 2 || usuarioAtualizacao != usuarioID {
		t.Fatalf("auditoria: version=%d usuario=%q erro=%v", version, usuarioAtualizacao, err)
	}

	cases := []struct {
		name, body, token, ifMatch, targetID string
		status                               int
	}{
		{"versão desatualizada", `{"nome":"Outro"}`, token, "1", id, http.StatusPreconditionFailed},
		{"sem If-Match", `{"nome":"Outro"}`, token, "", id, http.StatusPreconditionRequired},
		{"campo imutável", `{"ativo":false}`, token, "2", id, http.StatusBadRequest},
		{"nome duplicado", `{"nome":" SERVIÇO DUPLICADO INTEGRAÇÃO "}`, token, "2", id, http.StatusConflict},
		{"valor inválido", `{"valor":-1}`, token, "2", id, http.StatusBadRequest},
		{"não encontrado", `{"nome":"Outro"}`, token, "1", "4f000000-0000-0000-0000-0000000000ff", http.StatusNotFound},
		{"sem autenticação", `{"nome":"Outro"}`, "", "2", id, http.StatusUnauthorized},
		{"sem permissão", `{"nome":"Outro"}`, semEscopo, "2", id, http.StatusForbidden},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := executarPATCH(handler, test.targetID, test.body, test.token, test.ifMatch)
			if got.Code != test.status {
				t.Fatalf("status %d: %s", got.Code, got.Body.String())
			}
		})
	}
}

func executarPATCH(handler http.Handler, id, body, token, ifMatch string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPatch, "/servicos/"+id, strings.NewReader(body))
	request.SetPathValue("servicoId", id)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if ifMatch != "" {
		request.Header.Set("If-Match", ifMatch)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
