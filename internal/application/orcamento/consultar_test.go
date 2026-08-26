package orcamento

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type repositoryStub struct {
	osID, clienteID string
	result          domain.Consulta
	err             error
}

func (stub *repositoryStub) Consultar(_ context.Context, osID, clienteID string) (domain.Consulta, error) {
	stub.osID, stub.clienteID = osID, clienteID
	return stub.result, stub.err
}

func TestConsultarDelegaAoRepositorio(t *testing.T) {
	stub := &repositoryStub{result: domain.Consulta{OrdemServicoID: "os"}}
	result, err := NewConsultar(stub).Execute(context.Background(), "os", "cliente")
	if err != nil || result.OrdemServicoID != "os" || stub.osID != "os" || stub.clienteID != "cliente" {
		t.Fatalf("result=%+v err=%v stub=%+v", result, err, stub)
	}
}
