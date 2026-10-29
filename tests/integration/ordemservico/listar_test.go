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

func TestListarOrdensDeServico(t *testing.T) {
	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Skip("banco indisponível")
	}
	t.Cleanup(db.Close)
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	clienteID, veiculoID := id("f1000000-0000-0000-0000-"), id("f2000000-0000-0000-0000-")
	documento := cpfValido(suffix % 1000000000)
	placa := placaMercosul("LST", suffix)

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Listagem',$2,'CPF','11999999999')", clienteID, documento); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Fiat','Uno',2021)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}

	// Uma OS por status. As datas crescem na ordem de prioridade, ao contrario do que a listagem
	// devolve: se a ordenacao fosse por data, a OS em execucao viria por ultimo.
	base := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	ordens := []struct {
		id, status string
		criadaEm   time.Time
	}{
		{id("f3000000-0000-0000-0000-"), "RECEBIDA", base},
		{id("f3100000-0000-0000-0000-"), "RECEBIDA", base.Add(time.Minute)},
		{id("f3200000-0000-0000-0000-"), "EM_DIAGNOSTICO", base.Add(2 * time.Minute)},
		{id("f3300000-0000-0000-0000-"), "AGUARDANDO_APROVACAO", base.Add(3 * time.Minute)},
		{id("f3400000-0000-0000-0000-"), "AGUARDANDO_RECURSOS", base.Add(4 * time.Minute)},
		{id("f3500000-0000-0000-0000-"), "AGUARDANDO_EXECUCAO", base.Add(5 * time.Minute)},
		{id("f3600000-0000-0000-0000-"), "EM_EXECUCAO", base.Add(6 * time.Minute)},
		{id("f3700000-0000-0000-0000-"), "FINALIZADA", base},
		{id("f3800000-0000-0000-0000-"), "ENTREGUE", base},
		{id("f3900000-0000-0000-0000-"), "CANCELADA", base},
	}
	for _, ordem := range ordens {
		if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status,criada_em) VALUES ($1,$2,$3,$4,$5,$6)",
			ordem.id, clienteID, veiculoID, placa, ordem.status, ordem.criadaEm); err != nil {
			t.Fatal(err)
		}
	}
	ordemEsperada := []string{ordens[6].id, ordens[5].id, ordens[4].id, ordens[3].id, ordens[2].id, ordens[0].id, ordens[1].id}

	clientePaginacaoID, veiculoPaginacaoID := id("f5000000-0000-0000-0000-"), id("f6000000-0000-0000-0000-")
	documentoPaginacao := cpfValido((suffix + 1) % 1000000000)
	placaPaginacao := placaMercosul("PGN", suffix+1)
	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Paginacao',$2,'CPF','11999999998')", clientePaginacaoID, documentoPaginacao); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'VW','Gol',2020)", veiculoPaginacaoID, clientePaginacaoID, placaPaginacao); err != nil {
		t.Fatal(err)
	}
	mesmoInstante := time.Now().UTC().Truncate(time.Second)
	osPagina1ID, osPagina2ID := id("f7000000-0000-0000-0000-"), id("f8000000-0000-0000-0000-")
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status,criada_em) VALUES ($1,$3,$4,$5,'RECEBIDA',$6),($2,$3,$4,$5,'RECEBIDA',$6)",
		osPagina1ID, osPagina2ID, clientePaginacaoID, veiculoPaginacaoID, placaPaginacao, mesmoInstante); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE veiculo_id IN ($1,$2)", veiculoID, veiculoPaginacaoID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id IN ($1,$2)", veiculoID, veiculoPaginacaoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id IN ($1,$2)", clienteID, clientePaginacaoID)
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

	mux := http.NewServeMux()
	mux.Handle("GET /ordens-servico", segurancaPresentation.RequireScope(jwt, "os:ler",
		presentation.NewListarHandler(application.NewListar(infrastructure.NewPostgresRepository(db)))))

	requisitar := func(query, tok string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/ordens-servico"+query, nil)
		if tok != "" {
			request.Header.Set("Authorization", "Bearer "+tok)
		}
		writer := httptest.NewRecorder()
		mux.ServeHTTP(writer, request)
		return writer
	}
	type resposta struct {
		Data []struct {
			OrdemServicoID string `json:"ordemServicoId"`
			Status         string `json:"status"`
			CriadaEm       string `json:"criadaEm"`
		} `json:"data"`
		TotalElementos int `json:"totalElementos"`
	}
	listar := func(query string) resposta {
		t.Helper()
		writer := requisitar(query, token)
		if writer.Code != http.StatusOK {
			t.Fatalf("%s: status=%d body=%s", query, writer.Code, writer.Body.String())
		}
		var corpo resposta
		if err := json.Unmarshal(writer.Body.Bytes(), &corpo); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		return corpo
	}
	ids := func(corpo resposta) []string {
		lista := make([]string, 0, len(corpo.Data))
		for _, item := range corpo.Data {
			lista = append(lista, item.OrdemServicoID)
		}
		return lista
	}

	if writer := requisitar("", ""); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", writer.Code)
	}
	if writer := requisitar("", semEscopo); writer.Code != http.StatusForbidden {
		t.Fatalf("sem escopo=%d", writer.Code)
	}
	for _, status := range []string{"INVALIDO", "FINALIZADA", "ENTREGUE", "CANCELADA"} {
		if writer := requisitar("?status="+status, token); writer.Code != http.StatusBadRequest {
			t.Fatalf("status %s=%d", status, writer.Code)
		}
	}
	if writer := requisitar("?documento=52998224725", token); writer.Code != http.StatusNotFound {
		t.Fatalf("documento inexistente=%d", writer.Code)
	}
	if writer := requisitar("?placa=ZZZ9Z99", token); writer.Code != http.StatusNotFound {
		t.Fatalf("placa inexistente=%d", writer.Code)
	}

	// Prioridade de status primeiro, mais antiga primeiro dentro do status, sem as encerradas.
	porPlaca := listar("?placa=" + placa)
	if porPlaca.TotalElementos != len(ordemEsperada) || fmt.Sprint(ids(porPlaca)) != fmt.Sprint(ordemEsperada) {
		t.Fatalf("ordem da listagem:\n obtida   %v\n esperada %v", ids(porPlaca), ordemEsperada)
	}
	// Cada item traz a data de abertura gravada, para a ordem por data ficar visivel na resposta.
	criadaEmPorID := make(map[string]time.Time, len(ordens))
	for _, ordem := range ordens {
		criadaEmPorID[ordem.id] = ordem.criadaEm
	}
	for _, item := range porPlaca.Data {
		obtida, errData := time.Parse(time.RFC3339, item.CriadaEm)
		if esperada := criadaEmPorID[item.OrdemServicoID]; errData != nil || !obtida.Equal(esperada) {
			t.Fatalf("criadaEm de %s: obtida %q, esperada %s", item.OrdemServicoID, item.CriadaEm, esperada.Format(time.RFC3339))
		}
	}
	var gravadas int
	if err = db.QueryRow(ctx, "SELECT COUNT(*) FROM ordem_servico WHERE veiculo_id = $1", veiculoID).Scan(&gravadas); err != nil || gravadas != len(ordens) {
		t.Fatalf("as encerradas devem continuar no banco: gravadas=%d erro=%v", gravadas, err)
	}

	if porDocumento := listar("?documento=" + documento); fmt.Sprint(ids(porDocumento)) != fmt.Sprint(ordemEsperada) {
		t.Fatalf("filtro por documento: %v", ids(porDocumento))
	}
	if recebidas := listar("?placa=" + placa + "&status=RECEBIDA"); fmt.Sprint(ids(recebidas)) != fmt.Sprint([]string{ordens[0].id, ordens[1].id}) {
		t.Fatalf("filtro combinado deveria trazer as recebidas, a mais antiga primeiro: %v", ids(recebidas))
	}

	// A pagina segue a mesma ordem: a segunda pagina de tamanho 3 comeca na quarta OS.
	segundaPagina := listar("?placa=" + placa + "&tamanho=3&pagina=1")
	if segundaPagina.TotalElementos != len(ordemEsperada) || fmt.Sprint(ids(segundaPagina)) != fmt.Sprint(ordemEsperada[3:6]) {
		t.Fatalf("segunda pagina: %v, esperado %v", ids(segundaPagina), ordemEsperada[3:6])
	}

	// Mesmo status e mesmo instante: o id desempata e a paginacao nao repete OS.
	primeira := listar("?placa=" + placaPaginacao + "&status=RECEBIDA&tamanho=1&pagina=0")
	segunda := listar("?placa=" + placaPaginacao + "&status=RECEBIDA&tamanho=1&pagina=1")
	if len(primeira.Data) != 1 || len(segunda.Data) != 1 || primeira.Data[0].OrdemServicoID == segunda.Data[0].OrdemServicoID {
		t.Fatalf("paginacao instavel: primeira=%v segunda=%v", ids(primeira), ids(segunda))
	}
}

// cpfValido gera um CPF sintaticamente valido (digitos verificadores corretos) a partir de uma
// base numerica, para exercitar filtros que validam o documento antes de consultar o banco.
func cpfValido(base int64) string {
	digitos := fmt.Sprintf("%09d", base%1000000000)
	calcular := func(pesoInicial int) int {
		soma := 0
		peso := pesoInicial
		for _, c := range digitos {
			soma += int(c-'0') * peso
			peso--
		}
		resto := soma % 11
		if resto < 2 {
			return 0
		}
		return 11 - resto
	}
	primeiro := calcular(10)
	digitos += fmt.Sprintf("%d", primeiro)
	segundo := calcular(11)
	return digitos + fmt.Sprintf("%d", segundo)
}
