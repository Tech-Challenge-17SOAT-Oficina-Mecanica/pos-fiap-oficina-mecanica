package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	securityInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	securityPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

func TestRecusarOrcamento(t *testing.T) {
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

	suffix := fmt.Sprintf("%012d", time.Now().UnixNano()%1000000000000)
	clienteID := "20000000-0000-0000-0000-" + suffix
	osPrincipalID := "70000000-0000-0000-0000-" + suffix
	osComplementarID := "70000000-0000-0000-0001-" + suffix
	orcamentoPrincipalID := "74000000-0000-0000-0000-" + suffix
	orcamentoComplementarID := "74000000-0000-0000-0001-" + suffix
	orcamentoAprovadoID := "74000000-0000-0000-0002-" + suffix
	orcamentoComplementarAprovadoOSID := "70000000-0000-0000-0002-" + suffix

	if _, err = db.Exec(ctx, `INSERT INTO cliente (id, nome, documento, tipo_documento, telefone) VALUES ($1, 'Teste Recusa', $2, 'CPF', '11999990000')`, clienteID, suffix[:11]); err != nil {
		t.Fatal(err)
	}
	var veiculoID string
	if err = db.QueryRow(ctx, `INSERT INTO veiculo (cliente_id, placa, marca, modelo, ano) VALUES ($1, $2, 'Fiat', 'Uno', 2020) RETURNING id`, clienteID, "TST"+suffix[:4]).Scan(&veiculoID); err != nil {
		t.Fatal(err)
	}
	criarOS := func(id, status string) {
		if _, err := db.Exec(ctx, `INSERT INTO ordem_servico (id, cliente_id, veiculo_id, placa_veiculo, status) VALUES ($1, $2, $3, 'TST0000', $4)`, id, clienteID, veiculoID, status); err != nil {
			t.Fatal(err)
		}
	}
	criarOS(osPrincipalID, "AGUARDANDO_APROVACAO")
	criarOS(osComplementarID, "EM_EXECUCAO")
	criarOS(orcamentoComplementarAprovadoOSID, "EM_EXECUCAO")

	if _, err = db.Exec(ctx, `INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status) VALUES ($1, $2, 'PRINCIPAL', 'CRIADO')`, orcamentoPrincipalID, osPrincipalID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO orcamento (id, ordem_servico_id, tipo_orcamento, status) VALUES ($1, $2, 'PRINCIPAL', 'APROVADO')`, orcamentoAprovadoID, orcamentoComplementarAprovadoOSID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO orcamento (id, ordem_servico_id, orcamento_original_id, tipo_orcamento, status) VALUES ($1, $2, $3, 'COMPLEMENTAR', 'CRIADO')`, orcamentoComplementarID, orcamentoComplementarAprovadoOSID, orcamentoAprovadoID); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec(ctx, `DELETE FROM orcamento WHERE ordem_servico_id IN ($1, $2, $3)`, osPrincipalID, osComplementarID, orcamentoComplementarAprovadoOSID)
		_, _ = db.Exec(ctx, `DELETE FROM ordem_servico WHERE id IN ($1, $2, $3)`, osPrincipalID, osComplementarID, orcamentoComplementarAprovadoOSID)
		_, _ = db.Exec(ctx, `DELETE FROM veiculo WHERE id = $1`, veiculoID)
		_, _ = db.Exec(ctx, `DELETE FROM cliente WHERE id = $1`, clienteID)
	})

	jwt, err := securityInfrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, _ := jwt.Gerar("mecanico", []string{"orcamentos:decidir"})
	clienteToken, _ := jwt.GerarCliente(clienteID, osPrincipalID)
	clienteOutraOS, _ := jwt.GerarCliente("00000000-0000-0000-0000-000000000000", osPrincipalID)

	mux := http.NewServeMux()
	mux.Handle("POST /orcamentos/{orcamentoId}/recusar", securityPresentation.RequireScope(jwt, "orcamentos:decidir", presentation.NewRecusarHandler(application.NewRecusar(infrastructure.NewPostgresRepository(db)))))

	request := func(token, orcamentoID, body string) *httptest.ResponseRecorder {
		writer := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orcamentos/"+orcamentoID+"/recusar", strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		mux.ServeHTTP(writer, req)
		return writer
	}

	if writer := request("", orcamentoPrincipalID, ""); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", writer.Code)
	}
	if writer := request(clienteOutraOS, orcamentoPrincipalID, ""); writer.Code != http.StatusForbidden {
		t.Fatalf("cliente de outra OS=%d body=%s", writer.Code, writer.Body.String())
	}

	writer := request(clienteToken, orcamentoPrincipalID, `{"motivo":"valor acima do esperado"}`)
	if writer.Code != http.StatusOK || !strings.Contains(writer.Body.String(), `"statusOrdemServico":"CANCELADA"`) {
		t.Fatalf("recusar principal=%d body=%s", writer.Code, writer.Body.String())
	}

	if writer := request(mecanico, orcamentoPrincipalID, ""); writer.Code != http.StatusConflict {
		t.Fatalf("recusar ja decidido=%d body=%s", writer.Code, writer.Body.String())
	}

	writer = request(mecanico, orcamentoComplementarID, "")
	if writer.Code != http.StatusOK || !strings.Contains(writer.Body.String(), `"statusOrdemServico":"AGUARDANDO_EXECUCAO"`) {
		t.Fatalf("recusar complementar=%d body=%s", writer.Code, writer.Body.String())
	}

	if writer := request(mecanico, "00000000-0000-0000-0000-000000000000", ""); writer.Code != http.StatusNotFound {
		t.Fatalf("orcamento inexistente=%d", writer.Code)
	}
}
