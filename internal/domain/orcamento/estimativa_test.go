package orcamento

import "testing"

func TestDiasFila(t *testing.T) {
	casos := []struct {
		nome       string
		naFrente   int
		capacidade int
		esperado   int
	}{
		// Exemplo da documentação: 3 OS por dia, 3 à frente, a quarta espera 1 dia.
		{"exemplo da doc", 3, 3, 1},
		{"arredonda para cima", 4, 3, 2},
		{"fila vazia", 0, 3, 0},
		{"capacidade não configurada", 10, 0, 0},
		{"capacidade negativa", 10, -1, 0},
		{"exatamente um dia", 3, 3, 1},
		{"dois dias cheios", 6, 3, 2},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			dados := DadosEstimativa{OSNaFrente: caso.naFrente, CapacidadeDiaria: caso.capacidade}
			if obtido := dados.DiasFila(); obtido != caso.esperado {
				t.Fatalf("diasFila = %d, esperado %d", obtido, caso.esperado)
			}
		})
	}
}

func TestEstimativaPrincipal(t *testing.T) {
	dados := DadosEstimativa{PrazoItensDias: 5, TempoServicosDias: 2, OSNaFrente: 3, CapacidadeDiaria: 3}

	if obtido := EstimativaPrincipal(dados); obtido != 8 {
		t.Fatalf("estimativa = %d, esperado 8 (5 + 2 + 1)", obtido)
	}
}

// Oficina nova: nenhuma OS finalizada, nenhum item já comprado, capacidade não
// configurada. O cálculo entrega zero em vez de travar.
func TestEstimativaSemNenhumDado(t *testing.T) {
	if obtido := EstimativaPrincipal(DadosEstimativa{}); obtido != 0 {
		t.Fatalf("estimativa = %d, esperado 0 quando não há dado nenhum", obtido)
	}
}

func TestEstimativaComDadoParcial(t *testing.T) {
	casos := []struct {
		nome     string
		dados    DadosEstimativa
		esperado int
	}{
		{"só prazo de item", DadosEstimativa{PrazoItensDias: 5}, 5},
		{"só tempo de serviço", DadosEstimativa{TempoServicosDias: 2}, 2},
		{"fila sem capacidade configurada", DadosEstimativa{OSNaFrente: 9}, 0},
		{"prazo e tempo, sem fila", DadosEstimativa{PrazoItensDias: 5, TempoServicosDias: 2}, 7},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if obtido := EstimativaPrincipal(caso.dados); obtido != caso.esperado {
				t.Fatalf("estimativa = %d, esperado %d", obtido, caso.esperado)
			}
		})
	}
}

// O complementar parte da estimativa do principal e soma só o que ele acrescenta.
// A fila não entra de novo: a OS já está na fila uma vez.
func TestEstimativaComplementar(t *testing.T) {
	dados := DadosEstimativa{PrazoItensDias: 3, TempoServicosDias: 1, OSNaFrente: 9, CapacidadeDiaria: 3}

	if obtido := EstimativaComplementar(8, dados); obtido != 12 {
		t.Fatalf("estimativa = %d, esperado 12 (8 + 3 + 1, sem recontar fila)", obtido)
	}
}

func TestEstimativaComplementarSemAcrescimo(t *testing.T) {
	if obtido := EstimativaComplementar(8, DadosEstimativa{}); obtido != 8 {
		t.Fatalf("estimativa = %d, esperado 8; sem acréscimo mantém a do principal", obtido)
	}
}

// Entrada inconsistente não pode virar estimativa negativa.
func TestEstimativaNuncaNegativa(t *testing.T) {
	dados := DadosEstimativa{PrazoItensDias: -5, TempoServicosDias: -2}

	if obtido := EstimativaPrincipal(dados); obtido != 0 {
		t.Fatalf("estimativa = %d, esperado 0", obtido)
	}
	if obtido := EstimativaComplementar(-3, dados); obtido != 0 {
		t.Fatalf("estimativa complementar = %d, esperado 0", obtido)
	}
}
