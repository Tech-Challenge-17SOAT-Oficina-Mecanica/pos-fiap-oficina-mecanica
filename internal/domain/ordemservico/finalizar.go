package ordemservico

import (
	"errors"
	"time"
)

const StatusFinalizada = "FINALIZADA"

var ErrOSNaoEmExecucao = errors.New("ordem de servico nao esta em execucao")
var ErrServicosPendentes = errors.New("existem servicos autorizados pendentes de conclusao")
var ErrOrcamentoComplementarPendente = errors.New("existe orcamento complementar aguardando decisao do cliente")

// ItemPendenteBaixa e um item ainda reservado para a OS sem baixa de estoque registrada.
type ItemPendenteBaixa struct {
	ItemID     string
	Codigo     string
	Quantidade float64
}

// ErroReservasPendentes bloqueia a finalizacao enquanto existir reserva ativa sem baixa.
type ErroReservasPendentes struct{ Itens []ItemPendenteBaixa }

func (erro ErroReservasPendentes) Error() string {
	return "existem itens reservados sem baixa de estoque registrada"
}

// ResultadoFinalizacao e o retorno do caso de uso Finalizar.
type ResultadoFinalizacao struct {
	OrdemServicoID     string
	ClienteID          string
	Status             string
	DataFinalizacao    time.Time
	Observacoes        string
	NotificacaoEnviada bool
}
