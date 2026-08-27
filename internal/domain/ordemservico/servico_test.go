package ordemservico

import (
	"errors"
	"testing"
)

func TestNovosServicosCadastro(t *testing.T) {
	servicos, err := NovosServicosCadastro([]ServicoCadastro{{ServicoID: "  servico-1 ", Observacao: "  urgente "}})
	if err != nil || len(servicos) != 1 || servicos[0].ServicoID != "servico-1" || servicos[0].Observacao != "urgente" {
		t.Fatalf("servicos=%+v err=%v", servicos, err)
	}
	_, err = NovosServicosCadastro(nil)
	if !errors.Is(err, ErrServicosObrigatorios) {
		t.Fatalf("err=%v", err)
	}
	_, err = NovosServicosCadastro([]ServicoCadastro{{ServicoID: "servico"}, {ServicoID: "servico"}})
	if !errors.Is(err, ErrServicoDuplicado) {
		t.Fatalf("err=%v", err)
	}
}

func TestTipoOrcamentoParaServico(t *testing.T) {
	for _, test := range []struct {
		status, tipo string
		err          error
	}{
		{StatusEmDiagnostico, OrcamentoPrincipal, nil},
		{StatusEmExecucao, OrcamentoComplementar, nil},
		{"ENTREGUE", "", ErrStatusNaoPermiteServico},
	} {
		tipo, err := TipoOrcamentoParaServico(test.status)
		if tipo != test.tipo || !errors.Is(err, test.err) {
			t.Fatalf("status=%s tipo=%q err=%v", test.status, tipo, err)
		}
	}
}
