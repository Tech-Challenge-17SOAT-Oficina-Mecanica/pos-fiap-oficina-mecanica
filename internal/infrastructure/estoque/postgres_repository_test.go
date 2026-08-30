package estoque

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
)

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatorio")
	}
}

func TestHashRequisicao(t *testing.T) {
	input := application.RegistrarEntradaInput{
		DocumentoOrigem:      "NF-123",
		FornecedorID:         "fornecedor-id",
		PedidoCompraID:       "pedido-id",
		ConfirmarDivergencia: true,
		Itens: []application.ItemInput{{
			ItemID:        "item-id",
			Quantidade:    2,
			CustoUnitario: 10,
		}},
	}

	hash := hashRequisicao(input)
	if hash == "" || len(hash) != 64 {
		t.Fatalf("hash invalido: %q", hash)
	}
	if hash != hashRequisicao(input) {
		t.Fatal("hash deveria ser deterministico")
	}

	input.DocumentoOrigem = "NF-124"
	if hash == hashRequisicao(input) {
		t.Fatal("hash deveria mudar com a requisicao")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	if !isUniqueViolation(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("esperava unique violation")
	}
	if isUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Fatal("nao deveria aceitar outro codigo")
	}
	if isUniqueViolation(errors.New("db")) {
		t.Fatal("nao deveria aceitar erro comum")
	}
}
