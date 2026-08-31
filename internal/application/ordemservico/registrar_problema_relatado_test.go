package ordemservico

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type problemaRelatadoRepositoryFake struct {
	ordem    domain.OrdemDeServico
	err      error
	osID     string
	problema domain.ProblemaRelatado
}

func (fake *problemaRelatadoRepositoryFake) RegistrarProblemaRelatado(_ context.Context, osID string, problema domain.ProblemaRelatado) (domain.OrdemDeServico, error) {
	fake.osID, fake.problema = osID, problema
	return fake.ordem, fake.err
}

func TestRegistrarProblemaRelatado(t *testing.T) {
	repository := &problemaRelatadoRepositoryFake{ordem: domain.OrdemDeServico{ID: "os", Status: domain.StatusEmDiagnostico}}
	result, err := NewRegistrarProblemaRelatado(repository, nil, nil).Execute(context.Background(), RegistrarProblemaRelatadoInput{OrdemServicoID: "os", Descricao: "  Ruído ao frear  ", Observacoes: "  Há uma semana  "})
	if err != nil || result.Status != domain.StatusEmDiagnostico {
		t.Fatalf("resultado = %+v, erro = %v", result, err)
	}
	if repository.osID != "os" || repository.problema.Descricao != "Ruído ao frear" || repository.problema.Observacoes != "Há uma semana" {
		t.Fatalf("persistência: %q %+v", repository.osID, repository.problema)
	}
}

func TestRegistrarProblemaRelatadoRejeitaDescricaoVazia(t *testing.T) {
	_, err := NewRegistrarProblemaRelatado(&problemaRelatadoRepositoryFake{}, nil, nil).Execute(context.Background(), RegistrarProblemaRelatadoInput{Descricao: "  "})
	if !errors.Is(err, domain.ErrDescricaoProblemaRelatadoObrigatoria) {
		t.Fatalf("erro = %v", err)
	}
}

func TestRegistrarProblemaRelatadoPropagaErrosDoRepository(t *testing.T) {
	for _, expected := range []error{ErrOrdemServicoNaoEncontrada, ErrOrdemServicoForaDeRecebida, ErrProblemaRelatadoJaRegistrado, errors.New("db")} {
		t.Run(expected.Error(), func(t *testing.T) {
			_, err := NewRegistrarProblemaRelatado(&problemaRelatadoRepositoryFake{err: expected}, nil, nil).Execute(context.Background(), RegistrarProblemaRelatadoInput{Descricao: "Ruído"})
			if !errors.Is(err, expected) {
				t.Fatalf("erro = %v", err)
			}
		})
	}
}
