package ordemservico

import (
	"errors"
	"strings"
)

var ErrDescricaoProblemaRelatadoObrigatoria = errors.New("descrição do problema relatado é obrigatória")

type ProblemaRelatado struct {
	Descricao   string
	Observacoes string
}

func NovoProblemaRelatado(descricao, observacoes string) (ProblemaRelatado, error) {
	descricao = strings.TrimSpace(descricao)
	if descricao == "" {
		return ProblemaRelatado{}, ErrDescricaoProblemaRelatadoObrigatoria
	}
	return ProblemaRelatado{Descricao: descricao, Observacoes: strings.TrimSpace(observacoes)}, nil
}
