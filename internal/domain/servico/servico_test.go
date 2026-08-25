package servico_test

import (
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

func TestNovo_valido(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Troca de óleo", Valor: 150.0, TempoEstimadoMinutos: 60}
	s, err := domain.Novo(input)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if s.Nome != "Troca de óleo" {
		t.Errorf("nome incorreto: %q", s.Nome)
	}
	if !s.Ativo {
		t.Error("esperava ativo = true")
	}
	if s.Version != 1 {
		t.Errorf("esperava version 1, obteve %d", s.Version)
	}
}

func TestNovo_nomeVazio(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "  ", Valor: 100, TempoEstimadoMinutos: 30}
	_, err := domain.Novo(input)
	if err != domain.ErrNomeObrigatorio {
		t.Errorf("esperava ErrNomeObrigatorio, obteve %v", err)
	}
}

func TestNovo_valorNegativo(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Alinhamento", Valor: -1, TempoEstimadoMinutos: 30}
	_, err := domain.Novo(input)
	if err != domain.ErrValorInvalido {
		t.Errorf("esperava ErrValorInvalido, obteve %v", err)
	}
}

func TestNovo_tempoEstimadoZero(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Balanceamento", Valor: 80, TempoEstimadoMinutos: 0}
	_, err := domain.Novo(input)
	if err != domain.ErrTempoEstimadoObrigatorio {
		t.Errorf("esperava ErrTempoEstimadoObrigatorio, obteve %v", err)
	}
}

func TestNovo_tempoEstimadoNegativo(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Balanceamento", Valor: 80, TempoEstimadoMinutos: -5}
	_, err := domain.Novo(input)
	if err != domain.ErrTempoEstimadoObrigatorio {
		t.Errorf("esperava ErrTempoEstimadoObrigatorio, obteve %v", err)
	}
}

func TestNovo_valorZeroAceito(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Inspeção", Valor: 0, TempoEstimadoMinutos: 15}
	_, err := domain.Novo(input)
	if err != nil {
		t.Errorf("esperava nil, obteve %v", err)
	}
}

func TestNormalizarNome(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Troca de Óleo", "troca de oleo"},
		{"  Alinhamento  e  Balanceamento  ", "alinhamento e balanceamento"},
		{"Revisão Geral", "revisao geral"},
		{"TROCA DE ÓLEO", "troca de oleo"},
	}
	for _, c := range cases {
		got := domain.NormalizarNome(c.input)
		if got != c.expected {
			t.Errorf("NormalizarNome(%q) = %q, esperava %q", c.input, got, c.expected)
		}
	}
}

func TestNovo_nomeNormalizadoGerado(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Revisão Geral", Valor: 200, TempoEstimadoMinutos: 120}
	s, err := domain.Novo(input)
	if err != nil {
		t.Fatal(err)
	}
	if s.NomeNormalizado != "revisao geral" {
		t.Errorf("nome_normalizado incorreto: %q", s.NomeNormalizado)
	}
}

func TestAplicarAtualizacao_parcial(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Troca de óleo", Valor: 150, TempoEstimadoMinutos: 60}
	s, _ := domain.Novo(input)

	novoValor := 180.0
	updated, err := s.AplicarAtualizacao(domain.AtualizacaoInput{Valor: &novoValor})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Valor != 180 {
		t.Errorf("esperava valor 180, obteve %f", updated.Valor)
	}
	if updated.Nome != "Troca de óleo" {
		t.Errorf("nome não deveria ter mudado: %q", updated.Nome)
	}
}

func TestAplicarAtualizacao_nomeVazio(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Troca de óleo", Valor: 150, TempoEstimadoMinutos: 60}
	s, _ := domain.Novo(input)

	nome := "   "
	_, err := s.AplicarAtualizacao(domain.AtualizacaoInput{Nome: &nome})
	if err != domain.ErrNomeObrigatorio {
		t.Errorf("esperava ErrNomeObrigatorio, obteve %v", err)
	}
}

func TestAplicarAtualizacao_valorNegativo(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Troca de óleo", Valor: 150, TempoEstimadoMinutos: 60}
	s, _ := domain.Novo(input)

	val := -10.0
	_, err := s.AplicarAtualizacao(domain.AtualizacaoInput{Valor: &val})
	if err != domain.ErrValorInvalido {
		t.Errorf("esperava ErrValorInvalido, obteve %v", err)
	}
}

func TestAplicarAtualizacao_tempoEstimadoMenorQue1(t *testing.T) {
	input := domain.NovoCadastroInput{Nome: "Troca de óleo", Valor: 150, TempoEstimadoMinutos: 60}
	s, _ := domain.Novo(input)

	tempo := 0
	_, err := s.AplicarAtualizacao(domain.AtualizacaoInput{TempoEstimadoMinutos: &tempo})
	if err != domain.ErrTempoEstimadoObrigatorio {
		t.Errorf("esperava ErrTempoEstimadoObrigatorio, obteve %v", err)
	}
}
