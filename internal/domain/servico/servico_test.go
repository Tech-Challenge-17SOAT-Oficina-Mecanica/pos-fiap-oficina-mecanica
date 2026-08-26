package servico

import (
	"errors"
	"testing"
)

func TestNovo(t *testing.T) {
	t.Run("válido e normalizado", func(t *testing.T) {
		servico, err := Novo(NovoServicoInput{Nome: "  Tróca   de ÓLEO  ", Descricao: "  Completa  ", Valor: "0", TempoEstimadoMinutos: 60})
		if err != nil {
			t.Fatal(err)
		}
		if servico.Nome != "Tróca   de ÓLEO" || servico.NomeNormalizado != "troca de oleo" ||
			servico.Descricao != "Completa" || !servico.Ativo || servico.Version != 1 {
			t.Fatalf("serviço inesperado: %+v", servico)
		}
	})

	cases := []struct {
		name  string
		input NovoServicoInput
		err   error
	}{
		{"nome vazio", NovoServicoInput{Valor: "10", TempoEstimadoMinutos: 1}, ErrNomeObrigatorio},
		{"valor ausente", NovoServicoInput{Nome: "Teste", TempoEstimadoMinutos: 1}, ErrValorInvalido},
		{"valor negativo", NovoServicoInput{Nome: "Teste", Valor: "-0.01", TempoEstimadoMinutos: 1}, ErrValorInvalido},
		{"mais de duas casas", NovoServicoInput{Nome: "Teste", Valor: "10.001", TempoEstimadoMinutos: 1}, ErrValorInvalido},
		{"tempo ausente", NovoServicoInput{Nome: "Teste", Valor: "10"}, ErrTempoEstimadoInvalido},
		{"tempo negativo", NovoServicoInput{Nome: "Teste", Valor: "10", TempoEstimadoMinutos: -1}, ErrTempoEstimadoInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := Novo(test.input)
			if !errors.Is(err, test.err) {
				t.Fatalf("erro %v, esperado %v", err, test.err)
			}
		})
	}
}

func TestNormalizarNome(t *testing.T) {
	if got := NormalizarNome("  Revisão\tBÁSICA com Limpeza  "); got != "revisao basica com limpeza" {
		t.Fatalf("normalização: %q", got)
	}
}

func TestAtualizar(t *testing.T) {
	original := Servico{ID: "id", Codigo: "SER-000001", Nome: "Revisão", NomeNormalizado: "revisao",
		Descricao: "Original", Valor: "100.00", TempoEstimadoMinutos: 30, Ativo: true, Version: 2}

	t.Run("parcial preserva os demais campos", func(t *testing.T) {
		nome := "  Revisão   Completa "
		atualizado, err := original.Atualizar(Atualizacao{Nome: &nome})
		if err != nil || atualizado.ID != original.ID || atualizado.Codigo != original.Codigo ||
			atualizado.Descricao != original.Descricao || atualizado.Valor != original.Valor ||
			atualizado.NomeNormalizado != "revisao completa" || atualizado.Version != original.Version {
			t.Fatalf("atualizado: %+v, erro: %v", atualizado, err)
		}
	})

	t.Run("atualiza campos permitidos", func(t *testing.T) {
		descricao, valor, tempo := " Nova descrição ", "180.00", 75
		atualizado, err := original.Atualizar(Atualizacao{Descricao: &descricao, Valor: &valor, TempoEstimadoMinutos: &tempo})
		if err != nil || atualizado.Descricao != "Nova descrição" || atualizado.Valor != "180.00" || atualizado.TempoEstimadoMinutos != 75 {
			t.Fatalf("atualizado: %+v, erro: %v", atualizado, err)
		}
	})

	cases := []struct {
		name        string
		atualizacao Atualizacao
		err         error
	}{
		{"vazia", Atualizacao{}, errors.New("ao menos um campo deve ser informado")},
		{"nome vazio", Atualizacao{Nome: stringPointer(" ")}, ErrNomeObrigatorio},
		{"valor inválido", Atualizacao{Valor: stringPointer("-1")}, ErrValorInvalido},
		{"tempo inválido", Atualizacao{TempoEstimadoMinutos: intPointer(0)}, ErrTempoEstimadoInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := original.Atualizar(test.atualizacao)
			if err == nil || test.err != nil && err.Error() != test.err.Error() {
				t.Fatalf("erro: %v", err)
			}
		})
	}
}

func TestAlterarSituacao(t *testing.T) {
	ativo := Servico{ID: "id", Ativo: true}
	inativo, err := ativo.Desativar()
	if err != nil || inativo.Ativo {
		t.Fatalf("serviço: %+v, erro: %v", inativo, err)
	}
	if _, err := inativo.Desativar(); !errors.Is(err, ErrServicoJaInativo) {
		t.Fatalf("erro ao desativar novamente: %v", err)
	}
	reativado, err := inativo.Reativar()
	if err != nil || !reativado.Ativo {
		t.Fatalf("serviço: %+v, erro: %v", reativado, err)
	}
	if _, err := ativo.Reativar(); !errors.Is(err, ErrServicoJaAtivo) {
		t.Fatalf("erro ao reativar novamente: %v", err)
	}
}

func stringPointer(value string) *string { return &value }
func intPointer(value int) *int          { return &value }
