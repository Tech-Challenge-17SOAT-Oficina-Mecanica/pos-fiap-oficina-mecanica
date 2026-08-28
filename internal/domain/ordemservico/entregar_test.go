package ordemservico

import "testing"

func TestValidarEntrega(t *testing.T) {
	if err := ValidarEntrega(StatusFinalizada); err != nil {
		t.Fatalf("FINALIZADA deveria permitir entrega: %v", err)
	}
	for _, status := range []string{StatusRecebida, StatusEmDiagnostico, StatusEmExecucao, "CANCELADA"} {
		if err := ValidarEntrega(status); err != ErrOSNaoFinalizada {
			t.Errorf("status=%s erro=%v, esperado %v", status, err, ErrOSNaoFinalizada)
		}
	}
	if err := ValidarEntrega(StatusEntregue); err != ErrOSJaEntregue {
		t.Fatalf("status ENTREGUE erro=%v, esperado %v", err, ErrOSJaEntregue)
	}
}
