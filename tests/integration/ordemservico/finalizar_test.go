package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestFinalizarServico(t *testing.T) {
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
	clienteID, veiculoID, categoriaID := id("d1000000-0000-0000-0000-"), id("d2000000-0000-0000-0000-"), id("d3000000-0000-0000-0000-")
	osOK, osServicoPendente, osComplementarPendente, osReservaPendente, osForaDeExecucao := id("d4000000-0000-0000-0000-"), id("d5000000-0000-0000-0000-"), id("d6000000-0000-0000-0000-"), id("d7000000-0000-0000-0000-"), id("d8000000-0000-0000-0000-")
	servicoID, itemID, osItemID := id("d9000000-0000-0000-0000-"), id("da000000-0000-0000-0000-"), id("db000000-0000-0000-0000-")
	placa := placaMercosul("FIN", suffix)
	codigoBase := suffix & 0xffffff

	if _, err = db.Exec(ctx, "INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)", categoriaID, fmt.Sprintf("Categoria %x", codigoBase)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES
		($1,$6,$7,$8,'EM_EXECUCAO'),($2,$6,$7,$8,'EM_EXECUCAO'),($3,$6,$7,$8,'EM_EXECUCAO'),($4,$6,$7,$8,'EM_EXECUCAO'),($5,$6,$7,$8,'RECEBIDA')`,
		osOK, osServicoPendente, osComplementarPendente, osReservaPendente, osForaDeExecucao, clienteID, veiculoID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO servico (id,codigo,nome,nome_normalizado,valor,tempo_estimado_minutos,ativo) VALUES ($1,$2,'Servico teste',$2,100,30,true)",
		servicoID, fmt.Sprintf("S%06x", codigoBase)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico_servico (id,ordem_servico_id,servico_id,descricao,valor_unitario,status) VALUES ($1,$2,$3,'Servico teste',100,'NECESSARIO')",
		id("dc000000-0000-0000-0000-"), osServicoPendente, servicoID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO orcamento (id,ordem_servico_id,tipo_orcamento,status) VALUES ($1,$2,'COMPLEMENTAR','CRIADO')",
		id("dd000000-0000-0000-0000-"), osComplementarPendente); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO item_estoque (id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado) VALUES
		($1,$2,'PECA',$3,'Peca teste','Peca teste','peca teste','UN',10,3)`, itemID, categoriaID, fmt.Sprintf("PEC-%06x", codigoBase)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico_item (id,ordem_servico_id,item_estoque_id,quantidade_necessaria,valor_unitario) VALUES ($1,$2,$3,3,10)",
		osItemID, osReservaPendente, itemID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO reserva_estoque (id,ordem_servico_item_id,item_estoque_id,quantidade,status) VALUES ($1,$2,$3,3,'ATIVA')",
		id("de000000-0000-0000-0000-"), osItemID, itemID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id IN ($1,$2,$3,$4,$5)", osOK, osServicoPendente, osComplementarPendente, osReservaPendente, osForaDeExecucao)
		_, _ = db.Exec(ctx, "DELETE FROM reserva_estoque WHERE ordem_servico_item_id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id=$1", itemID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE ordem_servico_id=$1", osComplementarPendente)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_servico WHERE ordem_servico_id=$1", osServicoPendente)
		_, _ = db.Exec(ctx, "DELETE FROM servico WHERE id=$1", servicoID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2,$3,$4,$5)", osOK, osServicoPendente, osComplementarPendente, osReservaPendente, osForaDeExecucao)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id=$1", categoriaID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "os:escrever",
		presentation.NewFinalizarHandler(application.NewFinalizar(infrastructure.NewPostgresRepository(db), nil, nil)))

	requisitar := func(osID, corpo string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/finalizar", strings.NewReader(corpo))
		request.SetPathValue("osId", osID)
		request.Header.Set("Authorization", "Bearer "+token)
		writer := httptest.NewRecorder()
		handler.ServeHTTP(writer, request)
		return writer
	}

	if writer := requisitar("00000000-0000-0000-0000-000000000000", `{}`); writer.Code != http.StatusNotFound {
		t.Fatalf("os inexistente: %d", writer.Code)
	}
	if writer := requisitar(osForaDeExecucao, `{}`); writer.Code != http.StatusConflict {
		t.Fatalf("fora de execucao: %d %s", writer.Code, writer.Body.String())
	}
	if writer := requisitar(osServicoPendente, `{}`); writer.Code != http.StatusConflict {
		t.Fatalf("servico pendente: %d %s", writer.Code, writer.Body.String())
	}
	if writer := requisitar(osComplementarPendente, `{}`); writer.Code != http.StatusConflict {
		t.Fatalf("complementar pendente: %d %s", writer.Code, writer.Body.String())
	}
	writer := requisitar(osReservaPendente, `{}`)
	if writer.Code != http.StatusConflict || !strings.Contains(writer.Body.String(), itemID) {
		t.Fatalf("reserva pendente: %d %s", writer.Code, writer.Body.String())
	}

	writer = requisitar(osOK, `{"observacoes":"Servicos concluidos e veiculo testado"}`)
	if writer.Code != http.StatusOK {
		t.Fatalf("finalizacao: %d %s", writer.Code, writer.Body.String())
	}
	var resposta map[string]any
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta["status"] != "FINALIZADA" {
		t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
	}

	var status string
	var finalizadaEm *time.Time
	if err = db.QueryRow(ctx, "SELECT status, finalizada_em FROM ordem_servico WHERE id=$1", osOK).Scan(&status, &finalizadaEm); err != nil || status != "FINALIZADA" || finalizadaEm == nil {
		t.Fatalf("status=%q finalizadaEm=%v erro=%v", status, finalizadaEm, err)
	}
	var auditorias int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM auditoria_ordem_servico WHERE ordem_servico_id=$1 AND usuario_id=$2 AND tipo_evento='FINALIZACAO'", osOK, "90000000-0000-0000-0000-000000000001").Scan(&auditorias); err != nil || auditorias != 1 {
		t.Fatalf("auditorias=%d erro=%v", auditorias, err)
	}

	if writer = requisitar(osOK, `{}`); writer.Code != http.StatusConflict {
		t.Fatalf("ja finalizada: %d", writer.Code)
	}
}
