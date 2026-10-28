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

const (
	clienteAna     = "20000000-0000-0000-0000-000000000001"
	veiculoAna     = "30000000-0000-0000-0000-000000000001"
	veiculoOutro   = "30000000-0000-0000-0000-000000000002"
	clienteAusente = "20000000-0000-0000-0000-999999999999"
)

type criarResponse struct {
	OrdemServicoID string    `json:"ordemServicoId"`
	ClienteID      string    `json:"clienteId"`
	VeiculoID      string    `json:"veiculoId"`
	Status         string    `json:"status"`
	CriadaEm       time.Time `json:"criadaEm"`
}

func TestCriarOrdemServico(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "os:escrever", presentation.NewCriarHandler(application.NewCriar(infrastructure.NewPostgresRepository(db))))

	request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(`{"clienteId":"`+clienteAna+`","veiculoId":"`+veiculoAna+`"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var created criarResponse
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(ctx, `DELETE FROM ordem_servico WHERE id = $1`, created.OrdemServicoID)
	if created.ClienteID != clienteAna || created.VeiculoID != veiculoAna || created.Status != "RECEBIDA" || created.CriadaEm.IsZero() {
		t.Fatalf("resposta: %+v", created)
	}

	var placa, status string
	var orcamentos int
	if err := db.QueryRow(ctx, `SELECT placa_veiculo, status FROM ordem_servico WHERE id = $1`, created.OrdemServicoID).Scan(&placa, &status); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM orcamento WHERE ordem_servico_id = $1`, created.OrdemServicoID).Scan(&orcamentos); err != nil {
		t.Fatal(err)
	}
	if placa != "ABC1D23" || status != "RECEBIDA" || orcamentos != 0 {
		t.Fatalf("placa = %q, status = %q, orçamentos = %d", placa, status, orcamentos)
	}

	assertStatus(t, handler, token, `{"clienteId":"`+clienteAusente+`","veiculoId":"`+veiculoAna+`"}`, http.StatusNotFound)
	assertStatus(t, handler, token, `{"clienteId":"`+clienteAna+`","veiculoId":"`+veiculoOutro+`"}`, http.StatusConflict)
}

func TestAbrirOrdemServicoComServicosEPecas(t *testing.T) {
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
	clienteID, veiculoID, categoriaID := id("c1000000-0000-0000-0000-"), id("c2000000-0000-0000-0000-"), id("c3000000-0000-0000-0000-")
	servicoID, servicoInativoID := id("c4000000-0000-0000-0000-"), id("c5000000-0000-0000-0000-")
	pecaID, pecaSemValorID := id("c6000000-0000-0000-0000-"), id("c7000000-0000-0000-0000-")
	itemAusenteID := id("c8000000-0000-0000-0000-")
	placa := placaMercosul("ABR", suffix)
	codigo := suffix & 0xffffff

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Abertura',$2,'CPF','11999999999')", clienteID, cpfValido(suffix%1000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Fiat','Uno',2021)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO servico (id,codigo,nome,nome_normalizado,valor,tempo_estimado_minutos,ativo) VALUES
		($1,$2,'Troca de oleo',$2,100,30,true),($3,$4,'Servico inativo',$4,10,30,false)`,
		servicoID, fmt.Sprintf("A%06x1", codigo), servicoInativoID, fmt.Sprintf("A%06x2", codigo)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)`, categoriaID, fmt.Sprintf("Categoria Abertura %x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO item_estoque
		(id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado,preco_venda,custo_unitario,ativo)
		VALUES ($1,$2,'PECA',$3,'Filtro','Filtro de oleo',$4,'UN',10,0,45,30,true),
		       ($5,$2,'PECA',$6,'Sem valor','Peca sem valor',$7,'UN',10,0,NULL,30,true)`,
		pecaID, categoriaID, fmt.Sprintf("B%06x1", codigo), fmt.Sprintf("B%06x1", codigo),
		pecaSemValorID, fmt.Sprintf("B%06x2", codigo), fmt.Sprintf("B%06x2", codigo)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		const daOS = "SELECT id FROM ordem_servico WHERE veiculo_id = $1"
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id IN ("+daOS+")", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento_item WHERE orcamento_id IN (SELECT id FROM orcamento WHERE ordem_servico_id IN ("+daOS+"))", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE ordem_servico_id IN ("+daOS+")", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE ordem_servico_id IN ("+daOS+")", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE veiculo_id = $1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id IN ($1,$2)", pecaID, pecaSemValorID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id = $1", categoriaID)
		_, _ = db.Exec(ctx, "DELETE FROM servico WHERE id IN ($1,$2)", servicoID, servicoInativoID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id = $1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id = $1", clienteID)
	})

	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	// A auditoria do registro de pecas referencia o usuario, entao o token usa o mecanico do seed.
	token, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", []string{"os:escrever"})
	if err != nil {
		t.Fatal(err)
	}
	handler := segurancaPresentation.RequireScope(jwt, "os:escrever", presentation.NewCriarHandler(application.NewCriar(infrastructure.NewPostgresRepository(db))))
	abrir := func(servicos, pecas string) string {
		return `{"clienteId":"` + clienteID + `","veiculoId":"` + veiculoID + `","servicos":[` + servicos + `],"pecas":[` + pecas + `]}`
	}

	request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(abrir(
		`{"servicoId":"`+servicoID+`","observacao":"cliente pediu"}`, `{"pecaId":"`+pecaID+`","quantidade":2}`)))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	var created criarResponse
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil || created.OrdemServicoID == "" || created.Status != "RECEBIDA" {
		t.Fatalf("resposta = %s, erro = %v", response.Body.String(), err)
	}

	var tipoOrcamento, statusOrcamento string
	var servicos, pecas int
	var total, quantidadeNaOS float64
	if err := db.QueryRow(ctx, `
		SELECT o.tipo_orcamento, o.status,
		       COUNT(*) FILTER (WHERE oi.tipo_item = 'SERVICO'), COUNT(*) FILTER (WHERE oi.tipo_item = 'PECA'),
		       SUM(oi.valor_total)
		FROM orcamento o JOIN orcamento_item oi ON oi.orcamento_id = o.id
		WHERE o.ordem_servico_id = $1
		GROUP BY o.tipo_orcamento, o.status`, created.OrdemServicoID).Scan(&tipoOrcamento, &statusOrcamento, &servicos, &pecas, &total); err != nil {
		t.Fatal(err)
	}
	if tipoOrcamento != "PRINCIPAL" || statusOrcamento != "CRIADO" || servicos != 1 || pecas != 1 || total != 190 {
		t.Fatalf("orçamento %s/%s, serviços = %d, peças = %d, total = %.2f", tipoOrcamento, statusOrcamento, servicos, pecas, total)
	}
	if err := db.QueryRow(ctx, `SELECT quantidade_necessaria FROM ordem_servico_item WHERE ordem_servico_id = $1 AND item_estoque_id = $2`,
		created.OrdemServicoID, pecaID).Scan(&quantidadeNaOS); err != nil || quantidadeNaOS != 2 {
		t.Fatalf("quantidade na OS = %.0f, erro = %v", quantidadeNaOS, err)
	}

	// Cada abertura abaixo falha num item depois da OS ja inserida: a transacao precisa desfazer tudo.
	assertStatus(t, handler, token, abrir(`{"servicoId":"`+servicoID+`"}`, `{"pecaId":"`+itemAusenteID+`","quantidade":1}`), http.StatusNotFound)
	assertStatus(t, handler, token, abrir(`{"servicoId":"`+itemAusenteID+`"}`, `{"pecaId":"`+pecaID+`","quantidade":1}`), http.StatusNotFound)
	assertStatus(t, handler, token, abrir(`{"servicoId":"`+servicoInativoID+`"}`, ""), http.StatusConflict)
	assertStatus(t, handler, token, abrir(`{"servicoId":"`+servicoID+`"}`, `{"pecaId":"`+pecaSemValorID+`","quantidade":1}`), http.StatusConflict)
	assertStatus(t, handler, token, abrir("", `{"pecaId":"`+pecaID+`","quantidade":1.5}`), http.StatusBadRequest)

	var ordens, orcamentos int
	if err := db.QueryRow(ctx, `SELECT COUNT(*), (SELECT COUNT(*) FROM orcamento o JOIN ordem_servico os ON os.id = o.ordem_servico_id WHERE os.veiculo_id = $1)
		FROM ordem_servico WHERE veiculo_id = $1`, veiculoID).Scan(&ordens, &orcamentos); err != nil {
		t.Fatal(err)
	}
	if ordens != 1 || orcamentos != 1 {
		t.Fatalf("após as falhas: ordens = %d, orçamentos = %d (esperado só a abertura válida)", ordens, orcamentos)
	}
}

func assertStatus(t *testing.T, handler http.Handler, token, body string, expected int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/ordens-servico", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("status %d, esperado %d: %s", response.Code, expected, response.Body.String())
	}
}
