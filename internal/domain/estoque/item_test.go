package estoque

import "testing"

func TestQuantidadeCompativelComUnidade(t *testing.T) {
	tests := []struct {
		nome          string
		quantidade    float64
		unidadeMedida string
		querErro      bool
	}{
		{"UN aceita inteiro", 5, "UN", false},
		{"UN rejeita fracao", 5.5, "UN", true},
		{"L aceita ate 3 casas", 1.234, "L", false},
		{"L rejeita mais de 3 casas", 1.2345, "L", true},
		{"KG aceita inteiro", 2, "KG", false},
	}
	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			err := QuantidadeCompativelComUnidade(test.quantidade, test.unidadeMedida)
			if test.querErro && err == nil {
				t.Fatal("esperava erro")
			}
			if !test.querErro && err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
}

func TestQuantidadeValida(t *testing.T) {
	tests := []struct {
		nome       string
		tipo       string
		quantidade float64
		querErro   bool
	}{
		{"peca aceita inteiro", TipoPeca, 2, false},
		{"peca rejeita fracao", TipoPeca, 2.5, true},
		{"insumo aceita fracao", TipoInsumo, 2.5, false},
		{"rejeita zero", TipoInsumo, 0, true},
		{"rejeita negativo", TipoPeca, -1, true},
	}
	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			err := QuantidadeValida(test.tipo, test.quantidade)
			if test.querErro && err == nil {
				t.Fatal("esperava erro")
			}
			if !test.querErro && err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
}
