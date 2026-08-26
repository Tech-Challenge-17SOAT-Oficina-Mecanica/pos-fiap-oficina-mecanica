package ordemservico

import (
	"context"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type repositoryStub struct {
	ordemServicoID string
	cadastro       domain.ProblemaCadastro
	resultado      ResultadoRegistroProblema
	err            error
}

func (stub *repositoryStub) RegistrarProblema(_ context.Context, ordemServicoID string, cadastro domain.ProblemaCadastro) (ResultadoRegistroProblema, error) {
	stub.ordemServicoID, stub.cadastro = ordemServicoID, cadastro
	return stub.resultado, stub.err
}

func TestRegistrarProblemaDelegaAoRepositorio(t *testing.T) {
	stub := &repositoryStub{resultado: ResultadoRegistroProblema{Problema: domain.Problema{ID: "problema"}}}
	resultado, err := NewRegistrarProblema(stub).Execute(context.Background(), "os", domain.ProblemaCadastro{Descricao: "problema"})
	if err != nil || resultado.Problema.ID != "problema" || stub.ordemServicoID != "os" || stub.cadastro.Descricao != "problema" {
		t.Fatalf("resultado=%+v erro=%v stub=%+v", resultado, err, stub)
	}
}
