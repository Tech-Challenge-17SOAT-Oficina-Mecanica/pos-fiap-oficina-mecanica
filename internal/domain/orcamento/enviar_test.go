package orcamento

import (
	"errors"
	"testing"
)

func comItens() []Item { return []Item{{ID: "i1", Quantidade: 1, ValorUnitario: 10}} }

// A maquina de estados: o principal sai de EM_DIAGNOSTICO, o complementar de EM_EXECUCAO.
func TestEnvioPermitido(t *testing.T) {
	casos := []struct {
		tipo     string
		status   string
		esperado bool
	}{
		{TipoPrincipal, OSStatusEmDiagnostico, true},
		{TipoPrincipal, OSStatusEmExecucao, false},
		{TipoPrincipal, "RECEBIDA", false},
		{TipoPrincipal, OSStatusAguardandoAprovacao, false},
		{TipoComplementar, OSStatusEmExecucao, true},
		{TipoComplementar, OSStatusEmDiagnostico, false},
		{TipoComplementar, OSStatusAguardandoAprovacao, false},
		{"QUALQUER", OSStatusEmDiagnostico, false},
	}

	for _, caso := range casos {
		t.Run(caso.tipo+"/"+caso.status, func(t *testing.T) {
			if obtido := EnvioPermitido(caso.tipo, caso.status); obtido != caso.esperado {
				t.Fatalf("permitido = %v, esperado %v", obtido, caso.esperado)
			}
		})
	}
}

func TestValidarParaEnvioAceitaOCaminhoFeliz(t *testing.T) {
	principal := Orcamento{Tipo: TipoPrincipal, Status: StatusCriado, Itens: comItens()}
	if err := principal.ValidarParaEnvio(OSStatusEmDiagnostico, true); err != nil {
		t.Fatalf("principal calculado em EM_DIAGNOSTICO deveria passar: %v", err)
	}

	complementar := Orcamento{Tipo: TipoComplementar, Status: StatusCriado, OriginalID: "p1", Itens: comItens()}
	if err := complementar.ValidarParaEnvio(OSStatusEmExecucao, true); err != nil {
		t.Fatalf("complementar calculado em EM_EXECUCAO deveria passar: %v", err)
	}
}

func TestValidarParaEnvioRecusa(t *testing.T) {
	casos := []struct {
		nome      string
		orcamento Orcamento
		statusOS  string
		calculado bool
		esperado  error
	}{
		{"ja aprovado", Orcamento{Tipo: TipoPrincipal, Status: StatusAprovado, Itens: comItens()},
			OSStatusEmDiagnostico, true, ErrStatusNaoCalculavel},
		{"ja recusado", Orcamento{Tipo: TipoPrincipal, Status: StatusRecusado, Itens: comItens()},
			OSStatusEmDiagnostico, true, ErrStatusNaoCalculavel},
		{"sem itens", Orcamento{Tipo: TipoPrincipal, Status: StatusCriado},
			OSStatusEmDiagnostico, true, ErrSemItens},
		{"nao calculado", Orcamento{Tipo: TipoPrincipal, Status: StatusCriado, Itens: comItens()},
			OSStatusEmDiagnostico, false, ErrOrcamentoNaoCalculado},
		{"ja enviado", Orcamento{Tipo: TipoPrincipal, Status: StatusCriado, Itens: comItens()},
			OSStatusAguardandoAprovacao, true, ErrOrcamentoJaEnviado},
		{"OS recem criada", Orcamento{Tipo: TipoPrincipal, Status: StatusCriado, Itens: comItens()},
			"RECEBIDA", true, ErrOSNaoPermiteEnvio},
		{"complementar em diagnostico", Orcamento{Tipo: TipoComplementar, Status: StatusCriado, Itens: comItens()},
			OSStatusEmDiagnostico, true, ErrOSNaoPermiteEnvio},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			err := caso.orcamento.ValidarParaEnvio(caso.statusOS, caso.calculado)
			if !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
		})
	}
}

// Enviar duas vezes seguidas pediria ao cliente que decidisse de novo sobre o mesmo
// orcamento — a segunda tentativa precisa ser recusada.
func TestEnviarDuasVezesEhRecusado(t *testing.T) {
	orcamento := Orcamento{Tipo: TipoPrincipal, Status: StatusCriado, Itens: comItens()}

	if err := orcamento.ValidarParaEnvio(OSStatusEmDiagnostico, true); err != nil {
		t.Fatal(err)
	}
	// depois do envio a OS esta em AGUARDANDO_APROVACAO
	if err := orcamento.ValidarParaEnvio(OSStatusAguardandoAprovacao, true); !errors.Is(err, ErrOrcamentoJaEnviado) {
		t.Fatalf("erro = %v, esperado %v", err, ErrOrcamentoJaEnviado)
	}
}
