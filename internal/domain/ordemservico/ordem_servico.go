package ordemservico

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusRecebida       = "RECEBIDA"
	StatusEmDiagnostico  = "EM_DIAGNOSTICO"
	TipoProblemaRelatado = "RELATADO"
)

var ErrDescricaoObrigatoria = errors.New("descrição é obrigatória")

type ProblemaRelatado struct {
	Descricao    string
	Observacoes  string
	RegistradoEm time.Time
}

type OrdemDeServico struct {
	ID                    string
	Status                string
	ProblemaRelatado      ProblemaRelatado
	DataInicioDiagnostico time.Time
}

func NovoProblemaRelatado(descricao, observacoes string) (ProblemaRelatado, error) {
	descricao = strings.TrimSpace(descricao)
	if descricao == "" {
		return ProblemaRelatado{}, ErrDescricaoObrigatoria
	}
	return ProblemaRelatado{Descricao: descricao, Observacoes: strings.TrimSpace(observacoes)}, nil
}
