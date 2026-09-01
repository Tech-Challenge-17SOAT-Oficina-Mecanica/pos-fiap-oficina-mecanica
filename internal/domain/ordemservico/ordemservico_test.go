package ordemservico

import "testing"

func TestPermiteRegistroDeItens(t *testing.T) {
	for _, status := range []string{"RECEBIDA", "EM_DIAGNOSTICO", "AGUARDANDO_RECURSOS", "AGUARDANDO_EXECUCAO", "EM_EXECUCAO"} {
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

// Com o orçamento já enviado, o cliente está decidindo sobre o que recebeu por e-mail.
// Aceitar item aqui faria ele aprovar uma coisa diferente da que leu.
func TestOrcamentoEnviadoNaoAceitaNovosItens(t *testing.T) {
	if PermiteRegistroDeItens("AGUARDANDO_APROVACAO") {
		t.Fatal("uma OS aguardando a decisão do cliente não pode receber itens novos")
	}
}
