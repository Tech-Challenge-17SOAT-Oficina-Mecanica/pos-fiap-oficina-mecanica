package ordemservico

import (
	"errors"
	"testing"
)

func TestNovoProblemaCadastro(t *testing.T) {
	cadastro, err := NovoProblemaCadastro("  Pastilhas gastas  ", "  verificar discos ")
	if err != nil {
		t.Fatal(err)
	}
	if cadastro.Descricao != "Pastilhas gastas" || cadastro.Observacoes != "verificar discos" {
		t.Fatalf("cadastro=%+v", cadastro)
	}
	_, err = NovoProblemaCadastro(" \t ", "")
	if !errors.Is(err, ErrDescricaoObrigatoria) {
		t.Fatalf("erro=%v", err)
	}
}

func TestTipoOrcamentoParaStatus(t *testing.T) {
	tests := []struct {
		status string
		tipo   string
		err    error
	}{
		{StatusEmDiagnostico, OrcamentoPrincipal, nil},
		{StatusEmExecucao, OrcamentoComplementar, nil},
		{"ENTREGUE", "", ErrStatusNaoPermiteProblema},
	}
	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			tipo, err := TipoOrcamentoParaStatus(test.status)
			if tipo != test.tipo || !errors.Is(err, test.err) {
				t.Fatalf("tipo=%q erro=%v", tipo, err)
			}
		})
	}
}
