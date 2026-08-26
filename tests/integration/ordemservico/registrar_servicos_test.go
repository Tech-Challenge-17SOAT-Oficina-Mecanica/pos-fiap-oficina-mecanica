package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

func TestRegistrarServicosNecessarios(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	clienteID, veiculoID := id("91000000-0000-0000-0000-"), id("92000000-0000-0000-0000-")
	osDiagnosticoID, osExecucaoID := id("93000000-0000-0000-0000-"), id("94000000-0000-0000-0000-")
	osFechadaID, osSemOrcamentoID, osRollbackID := id("95000000-0000-0000-0000-"), id("96000000-0000-0000-0000-"), id("97000000-0000-0000-0000-")
	orcamentoDiagnosticoID, orcamentoPrincipalID, orcamentoComplementarID, orcamentoRollbackID := id("98000000-0000-0000-0000-"), id("99000000-0000-0000-0000-"), id("9a000000-0000-0000-0000-"), id("9b000000-0000-0000-0000-")
	servico1ID, servico2ID, servicoInativoID := id("9c000000-0000-0000-0000-"), id("9d000000-0000-0000-0000-"), id("9e000000-0000-0000-0000-")
	plaque := fmt.Sprintf("SRV1A%02d", suffix%100)

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, plaque); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES
		($1,$2,$3,$4,'EM_DIAGNOSTICO'),($5,$2,$3,$4,'EM_EXECUCAO'),($6,$2,$3,$4,'ENTREGUE'),($7,$2,$3,$4,'EM_DIAGNOSTICO'),($8,$2,$3,$4,'EM_DIAGNOSTICO')`, osDiagnosticoID, clienteID, veiculoID, plaque, osExecucaoID, osFechadaID, osSemOrcamentoID, osRollbackID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO orcamento (id,ordem_servico_id,tipo_orcamento,status) VALUES
		($1,$2,'PRINCIPAL','CRIADO'),($3,$4,'PRINCIPAL','APROVADO'),($5,$4,'COMPLEMENTAR','CRIADO'),($6,$7,'PRINCIPAL','CRIADO')`, orcamentoDiagnosticoID, osDiagnosticoID, orcamentoPrincipalID, osExecucaoID, orcamentoComplementarID, orcamentoRollbackID, osRollbackID); err != nil {
		t.Fatal(err)
	}
	codigoBase := suffix & 0xffffff
	if _, err = db.Exec(ctx, `INSERT INTO servico (id,codigo,nome,nome_normalizado,valor,tempo_estimado_minutos,ativo) VALUES
		($1,$2,'Servico um',$2,100,30,true),($3,$4,'Servico dois',$4,50,30,true),($5,$6,'Servico inativo',$6,10,30,false)`, servico1ID, fmt.Sprintf("S%06x1", codigoBase), servico2ID, fmt.Sprintf("S%06x2", codigoBase), servicoInativoID, fmt.Sprintf("S%06x3", codigoBase)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM orcamento_item WHERE orcamento_id IN ($1,$2,$3,$4)", orcamentoDiagnosticoID, orcamentoPrincipalID, orcamentoComplementarID, orcamentoRollbackID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE id IN ($1,$2,$3,$4)", orcamentoDiagnosticoID, orcamentoPrincipalID, orcamentoComplementarID, orcamentoRollbackID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2,$3,$4,$5)", osDiagnosticoID, osExecucaoID, osFechadaID, osSemOrcamentoID, osRollbackID)
		_, _ = db.Exec(ctx, "DELETE FROM servico WHERE id IN ($1,$2,$3)", servico1ID, servico2ID, servicoInativoID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

	jwt, err := securityInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.Gerar("usuario", []string{"os:escrever"})
	semEscopo, _ := jwt.Gerar("usuario", []string{"os:ler"})
	mux := http.NewServeMux()
	mux.Handle("POST /ordens-servico/{osId}/servicos", securityPresentation.RequireScope(jwt, "os:escrever", presentation.NewRegistrarServicosHandler(application.NewRegistrarServicos(infrastructure.NewPostgresRepository(db)))))
	request := func(token, osID, body string) *httptest.ResponseRecorder {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/servicos", bytes.NewBufferString(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		mux.ServeHTTP(writer, req)
		return writer
	}
	body1 := `{"servicos":[{"servicoId":"` + servico1ID + `","observacao":" urgente "}]}`
	if writer := request("", osDiagnosticoID, body1); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", writer.Code)
	}
	if writer := request(semEscopo, osDiagnosticoID, body1); writer.Code != http.StatusForbidden {
		t.Fatalf("sem escopo=%d", writer.Code)
	}
	if writer := request(token, osDiagnosticoID, body1); writer.Code != http.StatusCreated {
		t.Fatalf("diagnostico=%d body=%s", writer.Code, writer.Body.String())
	}
	var descricao, observacao string
	var valor float64
	if err = db.QueryRow(ctx, "SELECT descricao, observacao, valor_total FROM orcamento_item WHERE orcamento_id=$1 AND servico_id=$2", orcamentoDiagnosticoID, servico1ID).Scan(&descricao, &observacao, &valor); err != nil || descricao != "Servico um" || observacao != "urgente" || valor != 100 {
		t.Fatalf("item=%q/%q/%.2f erro=%v", descricao, observacao, valor, err)
	}
	if writer := request(token, osDiagnosticoID, body1); writer.Code != http.StatusConflict {
		t.Fatalf("duplicado=%d", writer.Code)
	}
	if writer := request(token, osExecucaoID, `{"servicos":[{"servicoId":"`+servico2ID+`"}]}`); writer.Code != http.StatusCreated || !bytes.Contains(writer.Body.Bytes(), []byte(`"COMPLEMENTAR"`)) {
		t.Fatalf("execucao=%d body=%s", writer.Code, writer.Body.String())
	}
	if writer := request(token, osDiagnosticoID, `{"servicos":[{"servicoId":"`+servicoInativoID+`"}]}`); writer.Code != http.StatusConflict {
		t.Fatalf("inativo=%d", writer.Code)
	}
	if writer := request(token, osDiagnosticoID, `{"servicos":[{"servicoId":"90000000-0000-0000-0000-000000000099"}]}`); writer.Code != http.StatusNotFound {
		t.Fatalf("ausente=%d", writer.Code)
	}
	if writer := request(token, osFechadaID, body1); writer.Code != http.StatusConflict {
		t.Fatalf("fechada=%d", writer.Code)
	}
	if writer := request(token, osSemOrcamentoID, body1); writer.Code != http.StatusConflict {
		t.Fatalf("sem orcamento=%d", writer.Code)
	}
	if writer := request(token, osRollbackID, `{"servicos":[{"servicoId":"`+servico2ID+`"},{"servicoId":"90000000-0000-0000-0000-000000000099"}]}`); writer.Code != http.StatusNotFound {
		t.Fatalf("rollback=%d", writer.Code)
	}
	var itens int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM orcamento_item WHERE orcamento_id=$1", orcamentoRollbackID).Scan(&itens); err != nil || itens != 0 {
		t.Fatalf("itens apos rollback=%d erro=%v", itens, err)
	}
}
