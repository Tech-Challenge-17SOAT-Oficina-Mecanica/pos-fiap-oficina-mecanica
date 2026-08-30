package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestTempoExecucao(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	clienteID, veiculoID := id("a1000000-0000-0000-0000-"), id("a2000000-0000-0000-0000-")
	os120, os240, osEmExecucao, osSemInicio, osSemFim, osInconsistente := id("a3000000-0000-0000-0000-"), id("a4000000-0000-0000-0000-"), id("a5000000-0000-0000-0000-"), id("a6000000-0000-0000-0000-"), id("a7000000-0000-0000-0000-"), id("a8000000-0000-0000-0000-")
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Tempo',$2,'CPF','11999999999')", clienteID, cpfValido(suffix%1000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placaMercosul("TMP", suffix)); err != nil {
		t.Fatal(err)
	}
	inicio120 := time.Date(2026, time.August, 10, 9, 0, 0, 0, time.UTC)
	inicio240 := time.Date(2026, time.August, 11, 8, 0, 0, 0, time.UTC)
	if _, err = db.Exec(ctx, `INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status,iniciada_em,finalizada_em) VALUES
		($1,$6,$7,'TMP1A23','FINALIZADA',$8,$8 + INTERVAL '120 minutes'),
		($2,$6,$7,'TMP1A23','FINALIZADA',$9,$9 + INTERVAL '240 minutes'),
		($3,$6,$7,'TMP1A23','EM_EXECUCAO',$8,NULL),
		($4,$6,$7,'TMP1A23','FINALIZADA',NULL,$8 + INTERVAL '30 minutes'),
		($5,$6,$7,'TMP1A23','FINALIZADA',$8,NULL),
		($10,$6,$7,'TMP1A23','FINALIZADA',$8,$8 - INTERVAL '30 minutes')`, os120, os240, osEmExecucao, osSemInicio, osSemFim, clienteID, veiculoID, inicio120, inicio240, osInconsistente); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2,$3,$4,$5,$6)", os120, os240, osEmExecucao, osSemInicio, osSemFim, osInconsistente)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"os:ler"})
	if err != nil {
		t.Fatal(err)
	}
	semEscopo, err := jwt.Gerar("usuario", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	repositorio := infrastructure.NewPostgresRepository(db)
	mux := http.NewServeMux()
	mux.Handle("GET /ordens-servico/{osId}/tempo-execucao", segurancaPresentation.RequireScope(jwt, "os:ler", presentation.NewConsultarTempoExecucaoHandler(application.NewConsultarTempoExecucaoDaOS(repositorio))))
	mux.Handle("GET /ordens-servico/tempos-execucao", segurancaPresentation.RequireScope(jwt, "os:ler", presentation.NewListarTemposExecucaoHandler(application.NewConsultarTempoMedioExecucao(repositorio))))
	requisitar := func(caminho, tok string) *httptest.ResponseRecorder {
		requisicao := httptest.NewRequest(http.MethodGet, caminho, nil)
		if tok != "" {
			requisicao.Header.Set("Authorization", "Bearer "+tok)
		}
		resposta := httptest.NewRecorder()
		mux.ServeHTTP(resposta, requisicao)
		return resposta
	}

	if resposta := requisitar("/ordens-servico/tempos-execucao", ""); resposta.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", resposta.Code)
	}
	if resposta := requisitar("/ordens-servico/tempos-execucao", semEscopo); resposta.Code != http.StatusForbidden {
		t.Fatalf("sem escopo=%d", resposta.Code)
	}
	if resposta := requisitar("/ordens-servico/nao-e-uuid/tempo-execucao", token); resposta.Code != http.StatusBadRequest {
		t.Fatalf("osId invalido=%d", resposta.Code)
	}
	if resposta := requisitar("/ordens-servico/00000000-0000-0000-0000-000000000000/tempo-execucao", token); resposta.Code != http.StatusNotFound {
		t.Fatalf("os inexistente=%d", resposta.Code)
	}
	if resposta := requisitar("/ordens-servico/"+osEmExecucao+"/tempo-execucao", token); resposta.Code != http.StatusBadRequest {
		t.Fatalf("os incompleta=%d", resposta.Code)
	}
	if resposta := requisitar("/ordens-servico/"+osInconsistente+"/tempo-execucao", token); resposta.Code != http.StatusBadRequest {
		t.Fatalf("os inconsistente=%d", resposta.Code)
	}

	resposta := requisitar("/ordens-servico/"+os120+"/tempo-execucao", token)
	var individual struct {
		TempoExecucaoMinutos int `json:"tempoExecucaoMinutos"`
	}
	if resposta.Code != http.StatusOK || json.Unmarshal(resposta.Body.Bytes(), &individual) != nil || individual.TempoExecucaoMinutos != 120 {
		t.Fatalf("individual=%d %s", resposta.Code, resposta.Body.String())
	}
	resposta = requisitar("/ordens-servico/tempos-execucao?tamanho=1&pagina=1", token)
	var lista struct {
		TempoMedioExecucaoMinutos int               `json:"tempoMedioExecucaoMinutos"`
		TotalElementos            int               `json:"totalElementos"`
		TotalPaginas              int               `json:"totalPaginas"`
		Data                      []json.RawMessage `json:"data"`
	}
	if resposta.Code != http.StatusOK || json.Unmarshal(resposta.Body.Bytes(), &lista) != nil || lista.TempoMedioExecucaoMinutos != 180 || lista.TotalElementos != 2 || lista.TotalPaginas != 2 || len(lista.Data) != 1 {
		t.Fatalf("lista=%d %s", resposta.Code, resposta.Body.String())
	}
	resposta = requisitar("/ordens-servico/tempos-execucao?dataInicio=2026-08-11&dataFim=2026-08-11", token)
	if resposta.Code != http.StatusOK || json.Unmarshal(resposta.Body.Bytes(), &lista) != nil || lista.TempoMedioExecucaoMinutos != 240 || lista.TotalElementos != 1 {
		t.Fatalf("filtro=%d %s", resposta.Code, resposta.Body.String())
	}
	if resposta = requisitar("/ordens-servico/tempos-execucao?dataInicio=2026-08-12&dataFim=2026-08-10", token); resposta.Code != http.StatusBadRequest {
		t.Fatalf("periodo invalido=%d", resposta.Code)
	}
}
