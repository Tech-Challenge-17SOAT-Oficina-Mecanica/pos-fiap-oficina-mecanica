package orcamento

import (
	"context"
	"strings"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type recusarRepositoryStub struct {
	input  RecusarInput
	result domain.Decisao
	err    error
}

func (stub *recusarRepositoryStub) Recusar(_ context.Context, input RecusarInput) (domain.Decisao, error) {
	stub.input = input
	return stub.result, stub.err
}

func TestRecusarDelegaAoRepositorioComMotivoNormalizado(t *testing.T) {
	stub := &recusarRepositoryStub{result: domain.Decisao{OrcamentoID: "orcamento"}}
	result, err := NewRecusar(stub, nil, nil).Execute(context.Background(), RecusarInput{
		OrcamentoID: "orcamento",
		ClienteID:   "cliente",
		Motivo:      "  valor acima do esperado  ",
	})
	if err != nil || result.OrcamentoID != "orcamento" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if stub.input.Motivo != "valor acima do esperado" {
		t.Fatalf("motivo nao normalizado: %q", stub.input.Motivo)
	}
}

func TestRecusarRejeitaMotivoMuitoLongo(t *testing.T) {
	stub := &recusarRepositoryStub{}
	_, err := NewRecusar(stub, nil, nil).Execute(context.Background(), RecusarInput{
		OrcamentoID: "orcamento",
		Motivo:      strings.Repeat("a", 501),
	})
	if err != ErrMotivoInvalido {
		t.Fatalf("erro=%v, esperado %v", err, ErrMotivoInvalido)
	}
}
