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

func TestConsultarOrdemDeServico(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	clienteID, veiculoID, osID, orcamentoID, servicoID := id("e1000000-0000-0000-0000-"), id("e2000000-0000-0000-0000-"), id("e3000000-0000-0000-0000-"), id("e4000000-0000-0000-0000-"), id("e5000000-0000-0000-0000-")
	problemaID, orcamentoItemID, eventoID := id("e6000000-0000-0000-0000-"), id("e7000000-0000-0000-0000-"), id("e8000000-0000-0000-0000-")
	placa := fmt.Sprintf("CON1A%02d", suffix%100)
	codigoBase := suffix & 0xffffff

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Ana Consulta',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Fiat','Uno',2021)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES ($1,$2,$3,$4,'AGUARDANDO_APROVACAO')", osID, clienteID, veiculoID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO orcamento (id,ordem_servico_id,tipo_orcamento,status) VALUES ($1,$2,'PRINCIPAL','CRIADO')", orcamentoID, osID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO problema_ordem_servico (id,ordem_servico_id,orcamento_id,descricao) VALUES ($1,$2,$3,'Barulho ao frear')", problemaID, osID, orcamentoID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO servico (id,codigo,nome,nome_normalizado,valor,tempo_estimado_minutos,ativo) VALUES ($1,$2,'Troca de oleo',$2,150,60,true)", servicoID, fmt.Sprintf("S%06x1", codigoBase)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO orcamento_item (id,orcamento_id,servico_id,tipo_item,descricao,quantidade,valor_unitario,valor_total) VALUES ($1,$2,$3,'SERVICO','Troca de oleo',1,150,150)", orcamentoItemID, orcamentoID, servicoID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO auditoria_ordem_servico (id,ordem_servico_id,agregado,agregado_id,tipo_evento,dados,metadados,ocorrido_em)
		VALUES ($1,$2,'ORDEM_SERVICO',$2,'ORDEM_SERVICO_CRIADA','{}'::jsonb,'{}'::jsonb,CURRENT_TIMESTAMP)`, eventoID, osID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento_item WHERE orcamento_id=$1", orcamentoID)
		_, _ = db.Exec(ctx, "DELETE FROM problema_ordem_servico WHERE ordem_servico_id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE ordem_servico_id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM servico WHERE id=$1", servicoID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id=$1", osID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, err := jwt.Gerar("usuario", []string{"os:ler"})
	if err != nil {
		t.Fatal(err)
	}
	cliente, err := jwt.GerarCliente(clienteID, osID)
	if err != nil {
		t.Fatal(err)
	}
	clienteOutraOS, err := jwt.GerarCliente(clienteID, "20000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /ordens-servico/{osId}", segurancaPresentation.RequireAnyScope(jwt, []string{"os:ler", "orcamentos:ler"},
		presentation.NewConsultarHandler(application.NewConsultar(infrastructure.NewPostgresRepository(db)))))

	requisitar := func(id, token string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/ordens-servico/"+id, nil)
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		writer := httptest.NewRecorder()
		mux.ServeHTTP(writer, request)
		return writer
	}

	if writer := requisitar(osID, ""); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", writer.Code)
	}
	if writer := requisitar("00000000-0000-0000-0000-000000000000", mecanico); writer.Code != http.StatusNotFound {
		t.Fatalf("os inexistente=%d", writer.Code)
	}
	if writer := requisitar(osID, clienteOutraOS); writer.Code != http.StatusForbidden {
		t.Fatalf("cliente de outra os=%d", writer.Code)
	}

	writer := requisitar(osID, mecanico)
	if writer.Code != http.StatusOK {
		t.Fatalf("mecanico=%d body=%s", writer.Code, writer.Body.String())
	}
	var resposta map[string]any
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil {
		t.Fatalf("resposta invalida: %v", err)
	}
	if resposta["statusOrdemServico"] != "AGUARDANDO_APROVACAO" {
		t.Fatalf("statusOrdemServico=%v", resposta["statusOrdemServico"])
	}
	cliente_, _ := resposta["cliente"].(map[string]any)
	if cliente_["nome"] != "Ana Consulta" {
		t.Fatalf("cliente=%+v", cliente_)
	}
	problemas, _ := resposta["problemas"].([]any)
	if len(problemas) != 1 || problemas[0].(map[string]any)["orcamentoId"] != orcamentoID {
		t.Fatalf("problemas=%+v", problemas)
	}
	orcamentos, _ := resposta["orcamentos"].([]any)
	if len(orcamentos) != 1 {
		t.Fatalf("orcamentos=%+v", orcamentos)
	}
	itens, _ := orcamentos[0].(map[string]any)["itens"].([]any)
	if len(itens) != 1 {
		t.Fatalf("itens=%+v", itens)
	}
	if resposta["valorTotalGeral"] != 150.0 {
		t.Fatalf("valorTotalGeral=%v", resposta["valorTotalGeral"])
	}
	eventos, _ := resposta["eventos"].([]any)
	if len(eventos) != 1 {
		t.Fatalf("eventos=%+v", eventos)
	}

	if writer = requisitar(osID, cliente); writer.Code != http.StatusOK {
		t.Fatalf("cliente=%d body=%s", writer.Code, writer.Body.String())
	}
}
