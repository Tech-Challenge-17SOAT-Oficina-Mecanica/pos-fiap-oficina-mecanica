package ordemservico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type servicoRepositoryStub struct {
	ordemServicoID string
	servicos       []domain.ServicoCadastro
	resultado      ResultadoRegistroServicos
	err            error
}

func (stub *servicoRepositoryStub) RegistrarServicos(_ context.Context, ordemServicoID string, servicos []domain.ServicoCadastro) (ResultadoRegistroServicos, error) {
	stub.ordemServicoID, stub.servicos = ordemServicoID, servicos
	return stub.resultado, stub.err
}

func TestRegistrarServicosDelegaAoRepositorio(t *testing.T) {
	stub := &servicoRepositoryStub{resultado: ResultadoRegistroServicos{Orcamento: domain.Orcamento{ID: "orcamento"}}}
	resultado, err := NewRegistrarServicos(stub).Execute(context.Background(), "os", []domain.ServicoCadastro{{ServicoID: "servico"}})
	if err != nil || resultado.Orcamento.ID != "orcamento" || stub.ordemServicoID != "os" || len(stub.servicos) != 1 {
		t.Fatalf("resultado=%+v err=%v stub=%+v", resultado, err, stub)
	}
	stub.err = errors.New("falhou")
	_, err = NewRegistrarServicos(stub).Execute(context.Background(), "os", nil)
	if err == nil {
		t.Fatal("erro esperado")
	}
}
