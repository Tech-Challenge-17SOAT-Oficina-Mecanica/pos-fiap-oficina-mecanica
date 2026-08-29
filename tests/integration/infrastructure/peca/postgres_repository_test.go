package peca_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	pecaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

const (
	filtroCorreia = "correia"
	codigoFiltro  = "PEC-000001"
)

func repositorio(t *testing.T) pecaInfrastructure.PostgresRepository {
	t.Helper()
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL nao definida; teste de integracao ignorado")
	}
	pool, err := database.OpenPool()
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pecaInfrastructure.NewPostgresRepository(pool)
}

func TestBuscarPorFiltroPorCodigo(t *testing.T) {
	pecas, total, err := repositorio(t).BuscarPorFiltro(context.Background(),
		pecaApplication.Filtros{Codigo: codigoFiltro}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(pecas) != 1 {
		t.Fatalf("total = %d, itens = %d, esperado 1 e 1", total, len(pecas))
	}

	encontrada := pecas[0]
	if encontrada.Codigo != codigoFiltro || encontrada.Categoria == "" {
		t.Fatalf("peca ou categoria nao carregada: %+v", encontrada)
	}
	if encontrada.SaldoDisponivel() != encontrada.SaldoFisico-encontrada.SaldoReservado {
		t.Fatalf("saldo disponivel inconsistente: %+v", encontrada)
	}
	if encontrada.PrecoVenda == nil || *encontrada.PrecoVenda != "45.00" {
		t.Fatalf("precoVenda = %v, esperado 45.00 exato", encontrada.PrecoVenda)
	}
	if encontrada.Version < 1 {
		t.Fatalf("version nao carregada: %+v", encontrada)
	}
}

func TestBuscarPorFiltroDescricaoParcialEPedidoEmAberto(t *testing.T) {
	pecas, total, err := repositorio(t).BuscarPorFiltro(context.Background(),
		pecaApplication.Filtros{Descricao: filtroCorreia}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("total = %d, esperado 1", total)
	}
	if !pecas[0].PossuiPedidoEmAberto {
		t.Fatalf("PEC-000002 tem pedido ABERTO no seed: %+v", pecas[0])
	}
}

func TestBuscarPorFiltroNaoRetornaInsumo(t *testing.T) {
	_, total, err := repositorio(t).BuscarPorFiltro(context.Background(),
		pecaApplication.Filtros{Descricao: "oleo sintetico"}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Fatalf("total = %d; INS-000001 e insumo e nao pode aparecer em /estoque/pecas", total)
	}
}

func TestBuscarPorFiltroSomenteDisponiveis(t *testing.T) {
	repo := repositorio(t)

	_, comFiltro, err := repo.BuscarPorFiltro(context.Background(),
		pecaApplication.Filtros{Descricao: filtroCorreia, SomenteDisponiveis: true}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if comFiltro != 0 {
		t.Fatalf("PEC-000002 tem saldo zero e nao deveria passar em somenteDisponiveis, total = %d", comFiltro)
	}
}

func TestBuscarPorIDInexistente(t *testing.T) {
	_, err := repositorio(t).BuscarPorID(context.Background(), "00000000-0000-0000-0000-0000000000ff")
	if err != pecaApplication.ErrNaoEncontrada {
		t.Fatalf("erro = %v, esperado ErrNaoEncontrada", err)
	}
}

func TestBuscarPorIDCarregaPeca(t *testing.T) {
	encontrada, err := repositorio(t).BuscarPorID(context.Background(), "50000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if encontrada.Codigo != codigoFiltro || encontrada.Fabricante == nil || *encontrada.Fabricante != "Mann" {
		t.Fatalf("peca carregada incorretamente: %+v", encontrada)
	}
}

func TestCadastrarComFornecedor(t *testing.T) {
	fornecedorID := "60000000-0000-0000-0000-000000000001"
	fabricante := "Teste"
	precoVenda := "123.45"
	estoqueMinimo := int64(1)
	sufixo := time.Now().UnixNano()
	cadastro, err := peca.NovoCadastro(
		fmt.Sprintf("Bomba teste %d", sufixo),
		fmt.Sprintf("Bomba de agua teste fornecedor %d", sufixo),
		"10000000-0000-0000-0000-000000000001",
		&fabricante,
		&precoVenda,
		&estoqueMinimo,
		&fornecedorID,
	)
	if err != nil {
		t.Fatal(err)
	}

	cadastrada, err := repositorio(t).Cadastrar(context.Background(), cadastro)
	if err != nil {
		t.Fatal(err)
	}
	if cadastrada.FornecedorID == nil || *cadastrada.FornecedorID != fornecedorID {
		t.Fatalf("fornecedorId = %v, esperado %s", cadastrada.FornecedorID, fornecedorID)
	}
}
