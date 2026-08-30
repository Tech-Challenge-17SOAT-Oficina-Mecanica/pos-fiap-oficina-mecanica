package ordemservico

import (
	"errors"
	"time"
)

const StatusAguardandoExecucao = "AGUARDANDO_EXECUCAO"

var ErrOSNaoAptaParaExecucao = errors.New("ordem de servico nao esta apta para execucao")
var ErrOrcamentoNaoAprovado = errors.New("nao existe orcamento aprovado para a ordem de servico")
var ErrServicosNaoAutorizados = errors.New("nao existem servicos autorizados para execucao")
var ErrRecursosIndisponiveis = errors.New("existem pecas ou insumos necessarios sem reserva ativa")
var ErrMecanicoAutenticadoNaoEncontrado = errors.New("usuario autenticado nao possui mecanico cadastrado")

type ResultadoInicioExecucao struct {
	OrdemServicoID              string
	Status                      string
	MecanicoID                  string
	DataInicioExecucao          time.Time
	ItensBaixados               []ItemBaixadoInicioExecucao
	CustoTotalMateriaisBaixados float64
}

type ItemBaixadoInicioExecucao struct {
	ItemID              string
	Codigo              string
	Tipo                string
	UnidadeMedida       string
	QuantidadeBaixada   float64
	SaldoFisicoAtual    float64
	SaldoReservadoAtual float64
	CustoUnitario       float64
	CustoTotal          float64
}

func (ordem *OrdemDeServico) IniciarExecucao(dataInicio time.Time) error {
	if ordem.Status != StatusAguardandoExecucao {
		return ErrOSNaoAptaParaExecucao
	}
	ordem.Status = StatusEmExecucao
	ordem.DataInicioExecucao = &dataInicio
	return nil
}
