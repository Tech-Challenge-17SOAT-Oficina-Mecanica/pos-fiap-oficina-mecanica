package orcamento

import "errors"

// Status da OS relevantes para o envio do orcamento.
const (
	OSStatusEmDiagnostico = "EM_DIAGNOSTICO"
	OSStatusEmExecucao    = "EM_EXECUCAO"
)

var (
	ErrOrcamentoNaoCalculado = errors.New("calcule o orcamento antes de enviar ao cliente")
	ErrOrcamentoJaEnviado    = errors.New("orcamento ja foi enviado e aguarda a decisao do cliente")
	ErrOSNaoPermiteEnvio     = errors.New("a ordem de servico nao esta em um estado que permita enviar orcamento")
)

// EnvioPermitido diz se a OS pode receber o envio do orcamento, conforme a maquina de
// estados: o PRINCIPAL sai de EM_DIAGNOSTICO; o COMPLEMENTAR e criado com a OS em
// EM_EXECUCAO e a devolve para AGUARDANDO_APROVACAO ate a decisao do cliente.
func EnvioPermitido(tipoOrcamento, statusOS string) bool {
	if statusOS == OSStatusAguardandoAprovacao {
		return false // ja enviado, aguardando o cliente
	}
	switch tipoOrcamento {
	case TipoPrincipal:
		return statusOS == OSStatusEmDiagnostico
	case TipoComplementar:
		return statusOS == OSStatusEmExecucao
	default:
		return false
	}
}

// ValidarParaEnvio aplica as regras que independem de banco. O orcamento precisa estar
// CRIADO, ter sido calculado e ter itens: enviar ao cliente algo sem valor calculado
// seria pedir aprovacao de um numero que ainda nao existe.
func (orcamento Orcamento) ValidarParaEnvio(statusOS string, calculado bool) error {
	if orcamento.Status != StatusCriado {
		return ErrStatusNaoCalculavel
	}
	if len(orcamento.Itens) == 0 {
		return ErrSemItens
	}
	if !calculado {
		return ErrOrcamentoNaoCalculado
	}
	if statusOS == OSStatusAguardandoAprovacao {
		return ErrOrcamentoJaEnviado
	}
	if !EnvioPermitido(orcamento.Tipo, statusOS) {
		return ErrOSNaoPermiteEnvio
	}
	return nil
}
