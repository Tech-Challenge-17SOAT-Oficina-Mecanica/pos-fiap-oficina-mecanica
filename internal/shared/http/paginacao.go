package http

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	PaginaPadrao  = 0
	TamanhoPadrao = 20
	TamanhoMaximo = 50
)

var (
	ErrPaginaInvalida  = NovoErroCampo("pagina", "pagina deve ser maior ou igual a zero")
	ErrTamanhoInvalido = NovoErroCampo("tamanho", fmt.Sprintf("tamanho deve estar entre 1 e %d", TamanhoMaximo))
)

type Paginacao struct {
	Pagina  int
	Tamanho int
}

func LerPaginacao(query url.Values) (Paginacao, error) {
	pagina, err := lerInteiro(query.Get("pagina"), PaginaPadrao)
	if err != nil || pagina < 0 {
		return Paginacao{}, ErrPaginaInvalida
	}
	tamanho, err := lerInteiro(query.Get("tamanho"), TamanhoPadrao)
	if err != nil || tamanho < 1 || tamanho > TamanhoMaximo {
		return Paginacao{}, ErrTamanhoInvalido
	}
	return Paginacao{Pagina: pagina, Tamanho: tamanho}, nil
}

func (paginacao Paginacao) Limit() int { return paginacao.Tamanho }

func (paginacao Paginacao) Offset() int { return paginacao.Pagina * paginacao.Tamanho }

func lerInteiro(valor string, padrao int) (int, error) {
	if valor == "" {
		return padrao, nil
	}
	return strconv.Atoi(valor)
}
