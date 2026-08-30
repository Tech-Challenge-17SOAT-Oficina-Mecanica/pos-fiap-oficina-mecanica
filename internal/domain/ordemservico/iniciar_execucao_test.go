package ordemservico

import (
	"errors"
	"testing"
	"time"
)

func TestOrdemDeServicoIniciarExecucao(t *testing.T) {
	inicio := time.Now()
	ordem := OrdemDeServico{ID: "os-1", Status: StatusAguardandoExecucao}
	if err := ordem.IniciarExecucao(inicio); err != nil {
		t.Fatalf("inicio valido: %v", err)
	}
	if ordem.Status != StatusEmExecucao || ordem.DataInicioExecucao == nil || !ordem.DataInicioExecucao.Equal(inicio) {
		t.Fatalf("ordem apos inicio: %#v", ordem)
	}
	if err := ordem.IniciarExecucao(inicio); !errors.Is(err, ErrOSNaoAptaParaExecucao) {
		t.Fatalf("reinicio deveria falhar: %v", err)
	}
}
