package http

import (
	"encoding/json"
	stdhttp "net/http"
)

type Lista[T any] struct {
	Data           []T `json:"data"`
	Pagina         int `json:"pagina"`
	Tamanho        int `json:"tamanho"`
	TotalElementos int `json:"totalElementos"`
	TotalPaginas   int `json:"totalPaginas"`
}

func NovaLista[T any](itens []T, paginacao Paginacao, totalElementos int) Lista[T] {
	if itens == nil {
		itens = []T{}
	}
	return Lista[T]{
		Data:           itens,
		Pagina:         paginacao.Pagina,
		Tamanho:        paginacao.Tamanho,
		TotalElementos: totalElementos,
		TotalPaginas:   totalPaginas(totalElementos, paginacao.Tamanho),
	}
}

func WriteLista[T any](writer stdhttp.ResponseWriter, lista Lista[T]) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(stdhttp.StatusOK)
	_ = json.NewEncoder(writer).Encode(lista)
}

func totalPaginas(totalElementos, tamanho int) int {
	if totalElementos <= 0 || tamanho <= 0 {
		return 0
	}
	return (totalElementos + tamanho - 1) / tamanho
}
