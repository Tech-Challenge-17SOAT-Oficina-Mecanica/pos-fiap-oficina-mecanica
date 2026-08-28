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
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	clienteID, veiculoID, osRecebidaID, osCanceladaID := id("f1000000-0000-0000-0000-"), id("f2000000-0000-0000-0000-"), id("f3000000-0000-0000-0000-"), id("f4000000-0000-0000-0000-")
	documento := cpfValido(suffix % 1000000000)
	placa := fmt.Sprintf("LST%04d", suffix%10000)

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Listagem',$2,'CPF','11999999999')", clienteID, documento); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Fiat','Uno',2021)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES ($1,$3,$4,$5,'RECEBIDA'),($2,$3,$4,$5,'CANCELADA')",
		osRecebidaID, osCanceladaID, clienteID, veiculoID, placa); err != nil {
		t.Fatal(err)
	}
	clientePaginacaoID, veiculoPaginacaoID := id("f5000000-0000-0000-0000-"), id("f6000000-0000-0000-0000-")
	documentoPaginacao := cpfValido((suffix + 1) % 1000000000)
	placaPaginacao := fmt.Sprintf("PGN%04d", (suffix+1)%10000)
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
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2,$3,$4)", osRecebidaID, osCanceladaID, osPagina1ID, osPagina2ID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoPaginacaoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clientePaginacaoID)
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

	if writer := requisitar("", ""); writer.Code != http.StatusUnauthorized {
		t.Fatalf("sem token=%d", writer.Code)
	}
	if writer := requisitar("", semEscopo); writer.Code != http.StatusForbidden {
		t.Fatalf("sem escopo=%d", writer.Code)
	}
	if writer := requisitar("?status=INVALIDO", token); writer.Code != http.StatusBadRequest {
		t.Fatalf("status invalido=%d", writer.Code)
	}
	if writer := requisitar("?documento=52998224725", token); writer.Code != http.StatusNotFound {
		t.Fatalf("documento inexistente=%d", writer.Code)
	}
	if writer := requisitar("?placa=ZZZ9Z99", token); writer.Code != http.StatusNotFound {
		t.Fatalf("placa inexistente=%d", writer.Code)
	}

	writer := requisitar("?placa="+placa, token)
	if writer.Code != http.StatusOK {
		t.Fatalf("filtro por placa=%d body=%s", writer.Code, writer.Body.String())
	}
	var resposta struct {
		Data           []map[string]any `json:"data"`
		TotalElementos int              `json:"totalElementos"`
	}
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta.TotalElementos != 2 {
		t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
	}

	writer = requisitar("?placa="+placa+"&status=CANCELADA", token)
	if writer.Code != http.StatusOK {
		t.Fatalf("filtro combinado=%d body=%s", writer.Code, writer.Body.String())
	}
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta.TotalElementos != 1 || resposta.Data[0]["ordemServicoId"] != osCanceladaID {
		t.Fatalf("filtro combinado invalido: %s erro=%v", writer.Body.String(), err)
	}

	writer = requisitar("?documento="+documento, token)
	if writer.Code != http.StatusOK {
		t.Fatalf("filtro por documento=%d body=%s", writer.Code, writer.Body.String())
	}
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta.TotalElementos != 2 {
		t.Fatalf("filtro por documento invalido: %s erro=%v", writer.Body.String(), err)
	}

	writer = requisitar("?placa="+placaPaginacao+"&status=RECEBIDA&tamanho=1&pagina=0", token)
	if writer.Code != http.StatusOK {
		t.Fatalf("paginacao pagina0=%d body=%s", writer.Code, writer.Body.String())
	}
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || len(resposta.Data) != 1 {
		t.Fatalf("paginacao pagina0 invalida: %s erro=%v", writer.Body.String(), err)
	}
	primeiraID, _ := resposta.Data[0]["ordemServicoId"].(string)

	writer = requisitar("?placa="+placaPaginacao+"&status=RECEBIDA&tamanho=1&pagina=1", token)
	if writer.Code != http.StatusOK {
		t.Fatalf("paginacao pagina1=%d body=%s", writer.Code, writer.Body.String())
	}
	if err = json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || len(resposta.Data) != 1 {
		t.Fatalf("paginacao pagina1 invalida: %s erro=%v", writer.Body.String(), err)
	}
	segundaID, _ := resposta.Data[0]["ordemServicoId"].(string)
	if primeiraID == "" || segundaID == "" || primeiraID == segundaID {
		t.Fatalf("paginacao instavel: primeira=%s segunda=%s", primeiraID, segundaID)
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
