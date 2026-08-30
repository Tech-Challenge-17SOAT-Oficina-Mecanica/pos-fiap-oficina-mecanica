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

func TestIniciarExecucao(t *testing.T) {
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
	clienteID, veiculoID := id("e1000000-0000-0000-0000-"), id("e2000000-0000-0000-0000-")
	usuarioID, mecanicoID := id("e3000000-0000-0000-0000-"), id("e4000000-0000-0000-0000-")
	usuarioResponsavelID, mecanicoResponsavelID := id("e5000000-0000-0000-0000-"), id("e6000000-0000-0000-0000-")
	osValida, osComResponsavel, osSemOrcamento, osSemServico, osSemReserva, osJaIniciada := id("e7000000-0000-0000-0000-"), id("e8000000-0000-0000-0000-"), id("e9000000-0000-0000-0000-"), id("ea000000-0000-0000-0000-"), id("eb000000-0000-0000-0000-"), id("ec000000-0000-0000-0000-")
	itemID, osItemID, categoriaID := id("ee000000-0000-0000-0000-"), id("ef000000-0000-0000-0000-"), id("f0000000-0000-0000-0000-")

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Teste',$2,'CPF','11999999999')", clienteID, fmt.Sprintf("%011d", suffix%100000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placaMercosul("EXE", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO usuario (id,email,senha_hash) VALUES ($1,$2,'hash'),($3,$4,'hash')", usuarioID, "execucao-"+fmt.Sprintf("%x", suffix)+"@teste.local", usuarioResponsavelID, "responsavel-"+fmt.Sprintf("%x", suffix)+"@teste.local"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO mecanico (id,usuario_id,nome) VALUES ($1,$2,'Mecanico executor'),($3,$4,'Mecanico responsavel')", mecanicoID, usuarioID, mecanicoResponsavelID, usuarioResponsavelID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status,mecanico_responsavel_id) VALUES
		($1,$7,$8,'ABC1D23','AGUARDANDO_EXECUCAO',NULL),
		($2,$7,$8,'ABC1D23','AGUARDANDO_EXECUCAO',$9),
		($3,$7,$8,'ABC1D23','AGUARDANDO_EXECUCAO',NULL),
		($4,$7,$8,'ABC1D23','AGUARDANDO_EXECUCAO',NULL),
		($5,$7,$8,'ABC1D23','AGUARDANDO_EXECUCAO',NULL),
		($6,$7,$8,'ABC1D23','EM_EXECUCAO',NULL)`, osValida, osComResponsavel, osSemOrcamento, osSemServico, osSemReserva, osJaIniciada, clienteID, veiculoID, mecanicoResponsavelID); err != nil {
		t.Fatal(err)
	}
	for _, osID := range []string{osValida, osComResponsavel, osSemServico, osSemReserva, osJaIniciada} {
		if _, err = db.Exec(ctx, "INSERT INTO orcamento (id,ordem_servico_id,tipo_orcamento,status) VALUES (gen_random_uuid(),$1,'PRINCIPAL','APROVADO')", osID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec(ctx, `INSERT INTO orcamento_item (id,orcamento_id,tipo_item,descricao,quantidade,valor_unitario,valor_total)
		SELECT gen_random_uuid(), id, 'SERVICO', 'Servico autorizado', 1, 100, 100 FROM orcamento WHERE ordem_servico_id IN ($1,$2,$3,$4)`, osValida, osComResponsavel, osSemReserva, osJaIniciada); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)", categoriaID, "Categoria "+fmt.Sprintf("%x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO item_estoque (id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado) VALUES ($1,$2,'PECA',$3,'Peca','Peca','peca','UN',1,0)", itemID, categoriaID, "EXE-"+fmt.Sprintf("%06x", suffix&0xffffff)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico_item (id,ordem_servico_id,item_estoque_id,quantidade_necessaria,valor_unitario) VALUES ($1,$2,$3,1,10)", osItemID, osSemReserva, itemID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id IN ($1,$2,$3,$4,$5,$6)", osValida, osComResponsavel, osSemOrcamento, osSemServico, osSemReserva, osJaIniciada)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento_item WHERE orcamento_id IN (SELECT id FROM orcamento WHERE ordem_servico_id IN ($1,$2,$3,$4,$5,$6))", osValida, osComResponsavel, osSemOrcamento, osSemServico, osSemReserva, osJaIniciada)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE id=$1", osItemID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE ordem_servico_id IN ($1,$2,$3,$4,$5,$6)", osValida, osComResponsavel, osSemOrcamento, osSemServico, osSemReserva, osJaIniciada)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2,$3,$4,$5,$6)", osValida, osComResponsavel, osSemOrcamento, osSemServico, osSemReserva, osJaIniciada)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id=$1", itemID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id=$1", categoriaID)
		_, _ = db.Exec(ctx, "DELETE FROM mecanico WHERE id IN ($1,$2)", mecanicoID, mecanicoResponsavelID)
		_, _ = db.Exec(ctx, "DELETE FROM usuario WHERE id IN ($1,$2)", usuarioID, usuarioResponsavelID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar(usuarioID, []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	tokenSemEscopo, err := jwt.Gerar(usuarioID, []string{"os:ler"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "os:escrever", presentation.NewIniciarExecucaoHandler(application.NewIniciarExecucao(infrastructure.NewPostgresRepository(db))))
	requisitar := func(osID, tokenAutorizacao string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osID+"/execucao/iniciar", nil)
		request.SetPathValue("osId", osID)
		if tokenAutorizacao != "" {
			request.Header.Set("Authorization", "Bearer "+tokenAutorizacao)
		}
		writer := httptest.NewRecorder()
		handler.ServeHTTP(writer, request)
		return writer
	}

	if writer := requisitar("invalido", token); writer.Code != http.StatusBadRequest {
		t.Fatalf("osId invalido: %d", writer.Code)
	}
	if writer := requisitar(osValida, ""); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token: %d", writer.Code)
	}
	if writer := requisitar(osValida, tokenSemEscopo); writer.Code != http.StatusForbidden {
		t.Fatalf("sem escopo: %d", writer.Code)
	}
	requestComCorpo := httptest.NewRequest(http.MethodPost, "/ordens-servico/"+osValida+"/execucao/iniciar", strings.NewReader(`{}`))
	requestComCorpo.SetPathValue("osId", osValida)
	requestComCorpo.Header.Set("Authorization", "Bearer "+token)
	writerComCorpo := httptest.NewRecorder()
	handler.ServeHTTP(writerComCorpo, requestComCorpo)
	if writerComCorpo.Code != http.StatusBadRequest {
		t.Fatalf("corpo nao permitido: %d", writerComCorpo.Code)
	}
	if writer := requisitar("00000000-0000-0000-0000-000000000000", token); writer.Code != http.StatusNotFound {
		t.Fatalf("os inexistente: %d", writer.Code)
	}
	for _, osID := range []string{osSemOrcamento, osSemServico, osSemReserva, osJaIniciada} {
		if writer := requisitar(osID, token); writer.Code != http.StatusConflict {
			t.Fatalf("os %s deveria conflitar: %d %s", osID, writer.Code, writer.Body.String())
		}
	}

	writer := requisitar(osValida, token)
	if writer.Code != http.StatusOK {
		t.Fatalf("inicio valido: %d %s", writer.Code, writer.Body.String())
	}
	var resposta map[string]any
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta["status"] != "EM_EXECUCAO" || resposta["mecanicoId"] != mecanicoID {
		t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
	}
	var status, mecanicoPersistido string
	var iniciadaEm *time.Time
	if err = db.QueryRow(ctx, "SELECT status, mecanico_responsavel_id::text, iniciada_em FROM ordem_servico WHERE id=$1", osValida).Scan(&status, &mecanicoPersistido, &iniciadaEm); err != nil || status != "EM_EXECUCAO" || mecanicoPersistido != mecanicoID || iniciadaEm == nil {
		t.Fatalf("persistencia status=%s mecanico=%s iniciadaEm=%v erro=%v", status, mecanicoPersistido, iniciadaEm, err)
	}
	var auditorias int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM auditoria_ordem_servico WHERE ordem_servico_id=$1 AND usuario_id=$2 AND tipo_evento='EXECUCAO_INICIADA'", osValida, usuarioID).Scan(&auditorias); err != nil || auditorias != 1 {
		t.Fatalf("auditorias=%d erro=%v", auditorias, err)
	}
	writer = requisitar(osComResponsavel, token)
	if writer.Code != http.StatusOK || !jsonContainsMecanico(t, writer.Body.Bytes(), mecanicoResponsavelID) {
		t.Fatalf("mecanico existente: %d %s", writer.Code, writer.Body.String())
	}
	if err = db.QueryRow(ctx, "SELECT mecanico_responsavel_id::text FROM ordem_servico WHERE id=$1", osComResponsavel).Scan(&mecanicoPersistido); err != nil || mecanicoPersistido != mecanicoResponsavelID {
		t.Fatalf("mecanico foi sobrescrito: %s erro=%v", mecanicoPersistido, err)
	}
}

func jsonContainsMecanico(t *testing.T, body []byte, mecanicoID string) bool {
	t.Helper()
	var resposta map[string]any
	return json.Unmarshal(body, &resposta) == nil && resposta["mecanicoId"] == mecanicoID
}
