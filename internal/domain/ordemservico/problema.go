package ordemservico

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusEmDiagnostico   = "EM_DIAGNOSTICO"
	StatusEmExecucao      = "EM_EXECUCAO"
	OrcamentoPrincipal    = "PRINCIPAL"
	OrcamentoComplementar = "COMPLEMENTAR"
	OrcamentoCriado       = "CRIADO"
)

var ErrDescricaoObrigatoria = errors.New("descricao do problema e obrigatoria")
var ErrStatusNaoPermiteProblema = errors.New("status da ordem de servico nao permite registrar problema")
var ErrOrcamentoFechado = errors.New("orcamento nao esta aberto para receber problemas")
var ErrOrcamentoPrincipalNaoEncontrado = errors.New("orcamento principal nao encontrado para a ordem de servico")

type ProblemaCadastro struct {
	Descricao   string
	Observacoes string
}

type Problema struct {
	ID           string
	Descricao    string
	Observacoes  string
	RegistradoEm time.Time
}

type Orcamento struct {
	ID         string
	Tipo       string
	Status     string
	ValorTotal float64
}

func NovoProblemaCadastro(descricao, observacoes string) (ProblemaCadastro, error) {
	cadastro := ProblemaCadastro{Descricao: strings.TrimSpace(descricao), Observacoes: strings.TrimSpace(observacoes)}
	if cadastro.Descricao == "" {
		return ProblemaCadastro{}, ErrDescricaoObrigatoria
	}
	return cadastro, nil
}

func TipoOrcamentoParaStatus(status string) (string, error) {
	switch status {
	case StatusEmDiagnostico:
		return OrcamentoPrincipal, nil
	case StatusEmExecucao:
		return OrcamentoComplementar, nil
	default:
		return "", ErrStatusNaoPermiteProblema
	}
}
