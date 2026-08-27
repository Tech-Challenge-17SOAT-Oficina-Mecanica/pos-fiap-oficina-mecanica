package http

import (
	"errors"
	"net/url"
	"strconv"
	"testing"
)

func TestLerPaginacao(t *testing.T) {
	casos := []struct {
		nome     string
		query    url.Values
		esperada Paginacao
		erro     error
	}{
		{"sem parametros usa o padrao", url.Values{}, Paginacao{Pagina: PaginaPadrao, Tamanho: TamanhoPadrao}, nil},
		{"valores informados", url.Values{"pagina": {"2"}, "tamanho": {"7"}}, Paginacao{Pagina: 2, Tamanho: 7}, nil},
		{"tamanho no limite", url.Values{"tamanho": {strconv.Itoa(TamanhoMaximo)}}, Paginacao{Pagina: 0, Tamanho: TamanhoMaximo}, nil},
		{"pagina negativa", url.Values{"pagina": {"-1"}}, Paginacao{}, ErrPaginaInvalida},
		{"pagina nao numerica", url.Values{"pagina": {"abc"}}, Paginacao{}, ErrPaginaInvalida},
		{"tamanho zero", url.Values{"tamanho": {"0"}}, Paginacao{}, ErrTamanhoInvalido},
		{"tamanho acima do maximo", url.Values{"tamanho": {strconv.Itoa(TamanhoMaximo + 1)}}, Paginacao{}, ErrTamanhoInvalido},
		{"tamanho nao numerico", url.Values{"tamanho": {"abc"}}, Paginacao{}, ErrTamanhoInvalido},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			paginacao, err := LerPaginacao(caso.query)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro = %v, esperado %v", err, caso.erro)
			}
			if paginacao != caso.esperada {
				t.Fatalf("paginacao = %+v, esperada %+v", paginacao, caso.esperada)
			}
		})
	}
}

func TestPaginacaoLimitEOffset(t *testing.T) {
	paginacao := Paginacao{Pagina: 3, Tamanho: 20}
	if paginacao.Limit() != 20 {
		t.Fatalf("Limit() = %d, esperado 20", paginacao.Limit())
	}
	if paginacao.Offset() != 60 {
		t.Fatalf("Offset() = %d, esperado 60", paginacao.Offset())
	}
}
