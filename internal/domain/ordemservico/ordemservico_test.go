package ordemservico

import "testing"

func TestPermiteRegistroDeItens(t *testing.T) {
	for _, status := range []string{"RECEBIDA", "EM_DIAGNOSTICO", "AGUARDANDO_APROVACAO", "AGUARDANDO_EXECUCAO", "EM_EXECUCAO"} {
		if !PermiteRegistroDeItens(status) {
			t.Errorf("status %s deveria permitir itens", status)
		}
	}
	for _, status := range []string{"FINALIZADA", "ENTREGUE", "CANCELADA"} {
		if PermiteRegistroDeItens(status) {
			t.Errorf("status %s nao deveria permitir itens", status)
		}
	}
}
