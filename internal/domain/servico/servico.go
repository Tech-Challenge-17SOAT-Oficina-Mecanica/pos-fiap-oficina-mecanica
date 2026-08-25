package servico

import (
	"errors"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrNomeObrigatorio          = errors.New("nome é obrigatório")
	ErrValorInvalido            = errors.New("valor deve ser maior ou igual a zero")
	ErrTempoEstimadoObrigatorio = errors.New("tempoEstimadoMinutos é obrigatório e deve ser maior ou igual a 1")
	ErrServicoIDObrigatorio     = errors.New("servicoId é obrigatório")
	ErrCampoImutavel            = errors.New("campo imutável não pode ser alterado")
)

type Servico struct {
	ID                   string
	Codigo               string
	Nome                 string
	NomeNormalizado      string
	Descricao            string
	Valor                float64
	TempoEstimadoMinutos int
	Ativo                bool
	DataDesativacao      *time.Time
	UsuarioDesativacao   string
	DataCriacao          time.Time
	DataAtualizacao      *time.Time
	UsuarioAtualizacao   string
	Version              int
}

type NovoCadastroInput struct {
	Nome                 string
	Descricao            string
	Valor                float64
	TempoEstimadoMinutos int
}

type AtualizacaoInput struct {
	Nome                 *string
	Descricao            *string
	Valor                *float64
	TempoEstimadoMinutos *int
}

func Novo(input NovoCadastroInput) (Servico, error) {
	nome := strings.TrimSpace(input.Nome)
	if nome == "" {
		return Servico{}, ErrNomeObrigatorio
	}
	if input.Valor < 0 {
		return Servico{}, ErrValorInvalido
	}
	if input.TempoEstimadoMinutos < 1 {
		return Servico{}, ErrTempoEstimadoObrigatorio
	}
	return Servico{
		Nome:                 nome,
		NomeNormalizado:      NormalizarNome(nome),
		Descricao:            strings.TrimSpace(input.Descricao),
		Valor:                input.Valor,
		TempoEstimadoMinutos: input.TempoEstimadoMinutos,
		Ativo:                true,
		Version:              1,
	}, nil
}

func (s Servico) AplicarAtualizacao(input AtualizacaoInput) (Servico, error) {
	if input.Nome != nil {
		nome := strings.TrimSpace(*input.Nome)
		if nome == "" {
			return Servico{}, ErrNomeObrigatorio
		}
		s.Nome = nome
		s.NomeNormalizado = NormalizarNome(nome)
	}
	if input.Descricao != nil {
		s.Descricao = strings.TrimSpace(*input.Descricao)
	}
	if input.Valor != nil {
		if *input.Valor < 0 {
			return Servico{}, ErrValorInvalido
		}
		s.Valor = *input.Valor
	}
	if input.TempoEstimadoMinutos != nil {
		if *input.TempoEstimadoMinutos < 1 {
			return Servico{}, ErrTempoEstimadoObrigatorio
		}
		s.TempoEstimadoMinutos = *input.TempoEstimadoMinutos
	}
	return s, nil
}

// NormalizarNome remove acentos, normaliza espaços e converte para minúsculas.
func NormalizarNome(nome string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, nome)
	result = strings.ToLower(result)
	result = strings.Join(strings.Fields(result), " ")
	return result
}
