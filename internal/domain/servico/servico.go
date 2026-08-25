package servico

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var valorPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,2})?$`)

var (
	ErrNomeObrigatorio       = errors.New("nome é obrigatório")
	ErrValorInvalido         = errors.New("valor deve ser maior ou igual a zero")
	ErrTempoEstimadoInvalido = errors.New("tempoEstimadoMinutos deve ser maior ou igual a 1")
	ErrAtualizacaoVazia      = errors.New("ao menos um campo deve ser informado")
)

type NovoServicoInput struct {
	Nome                 string
	Descricao            string
	Valor                string
	TempoEstimadoMinutos int
}

type Servico struct {
	ID                   string
	Codigo               string
	Nome                 string
	NomeNormalizado      string
	Descricao            string
	Valor                string
	TempoEstimadoMinutos int
	Ativo                bool
	Version              int
	DataCriacao          time.Time
	DataAtualizacao      *time.Time
	UsuarioCriacao       string
	UsuarioAtualizacao   string
}

type Atualizacao struct {
	Nome                 *string
	Descricao            *string
	Valor                *string
	TempoEstimadoMinutos *int
}

func Novo(input NovoServicoInput) (Servico, error) {
	servico := Servico{
		Nome:                 strings.TrimSpace(input.Nome),
		Descricao:            strings.TrimSpace(input.Descricao),
		Valor:                input.Valor,
		TempoEstimadoMinutos: input.TempoEstimadoMinutos,
		Ativo:                true,
		Version:              1,
	}
	if servico.Nome == "" {
		return Servico{}, ErrNomeObrigatorio
	}
	if !valorPattern.MatchString(servico.Valor) {
		return Servico{}, ErrValorInvalido
	}
	if servico.TempoEstimadoMinutos < 1 {
		return Servico{}, ErrTempoEstimadoInvalido
	}
	servico.NomeNormalizado = NormalizarNome(servico.Nome)
	return servico, nil
}

func (servico Servico) Atualizar(atualizacao Atualizacao) (Servico, error) {
	if atualizacao.Nome == nil && atualizacao.Descricao == nil && atualizacao.Valor == nil && atualizacao.TempoEstimadoMinutos == nil {
		return Servico{}, ErrAtualizacaoVazia
	}
	if atualizacao.Nome != nil {
		servico.Nome = strings.TrimSpace(*atualizacao.Nome)
		if servico.Nome == "" {
			return Servico{}, ErrNomeObrigatorio
		}
		servico.NomeNormalizado = NormalizarNome(servico.Nome)
	}
	if atualizacao.Descricao != nil {
		servico.Descricao = strings.TrimSpace(*atualizacao.Descricao)
	}
	if atualizacao.Valor != nil {
		if !valorPattern.MatchString(*atualizacao.Valor) {
			return Servico{}, ErrValorInvalido
		}
		servico.Valor = *atualizacao.Valor
	}
	if atualizacao.TempoEstimadoMinutos != nil {
		if *atualizacao.TempoEstimadoMinutos < 1 {
			return Servico{}, ErrTempoEstimadoInvalido
		}
		servico.TempoEstimadoMinutos = *atualizacao.TempoEstimadoMinutos
	}
	return servico, nil
}

func NormalizarNome(nome string) string {
	nome = strings.ToLower(strings.TrimSpace(nome))
	nome = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'â', 'ã', 'ä':
			return 'a'
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		case 'í', 'ì', 'î', 'ï':
			return 'i'
		case 'ó', 'ò', 'ô', 'õ', 'ö':
			return 'o'
		case 'ú', 'ù', 'û', 'ü':
			return 'u'
		case 'ç':
			return 'c'
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, nome)
	return strings.Join(strings.Fields(nome), " ")
}
