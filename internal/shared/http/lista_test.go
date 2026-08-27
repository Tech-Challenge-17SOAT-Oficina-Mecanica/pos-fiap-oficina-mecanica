package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNovaListaCalculaTotalPaginas(t *testing.T) {
	casos := []struct {
		nome           string
		totalElementos int
		tamanho        int
		esperado       int
	}{
		{"exato", 40, 20, 2},
		{"com resto", 45, 20, 3},
		{"menos que uma pagina", 3, 20, 1},
		{"vazio", 0, 20, 0},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			lista := NovaLista([]string{}, Paginacao{Tamanho: caso.tamanho}, caso.totalElementos)
			if lista.TotalPaginas != caso.esperado {
				t.Fatalf("TotalPaginas = %d, esperado %d", lista.TotalPaginas, caso.esperado)
			}
		})
	}
}

func TestNovaListaTrocaSliceNiloPorVazio(t *testing.T) {
	var itens []string
	corpo, err := json.Marshal(NovaLista(itens, Paginacao{Pagina: 0, Tamanho: 20}, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(corpo), `"data":[]`) {
		t.Fatalf("data deveria ser [], recebido %s", corpo)
	}
}

func TestWriteLista(t *testing.T) {
	response := httptest.NewRecorder()
	WriteLista(response, NovaLista([]string{"PEC-000001"}, Paginacao{Pagina: 0, Tamanho: 20}, 1))

	if response.Code != 200 {
		t.Fatalf("status = %d, esperado 200", response.Code)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
	}

	var lista Lista[string]
	if err := json.Unmarshal(response.Body.Bytes(), &lista); err != nil {
		t.Fatal(err)
	}
	if len(lista.Data) != 1 || lista.TotalElementos != 1 || lista.TotalPaginas != 1 {
		t.Fatalf("envelope inválido: %+v", lista)
	}
}
