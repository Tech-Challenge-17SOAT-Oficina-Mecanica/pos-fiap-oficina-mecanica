package ordemservico

import (
	"slices"
	"testing"
)

func TestStatusListagemPorPrioridadeSegueOEnunciado(t *testing.T) {
	posicao := func(status string) int { return slices.Index(StatusListagemPorPrioridade, status) }
	ordemDoEnunciado := []string{StatusEmExecucao, "AGUARDANDO_APROVACAO", StatusEmDiagnostico, StatusRecebida}
	for i := 1; i < len(ordemDoEnunciado); i++ {
		if posicao(ordemDoEnunciado[i-1]) >= posicao(ordemDoEnunciado[i]) {
			t.Fatalf("%s deveria vir antes de %s: %v", ordemDoEnunciado[i-1], ordemDoEnunciado[i], StatusListagemPorPrioridade)
		}
	}
	for _, status := range []string{StatusAguardandoExecucao, "AGUARDANDO_RECURSOS"} {
		if posicao(status) <= posicao(StatusEmExecucao) || posicao(status) >= posicao("AGUARDANDO_APROVACAO") {
			t.Fatalf("%s deveria ficar entre execucao e aprovacao: %v", status, StatusListagemPorPrioridade)
		}
	}
}

func TestStatusListagemValidoExcluiOsEncerrados(t *testing.T) {
	for _, status := range StatusListagemPorPrioridade {
		if !StatusListagemValido(status) {
			t.Fatalf("%s deveria ser aceito", status)
		}
	}
	for _, status := range []string{StatusFinalizada, StatusEntregue, "CANCELADA", "", "INEXISTENTE"} {
		if StatusListagemValido(status) {
			t.Fatalf("%s nao deveria ser aceito", status)
		}
	}
}
