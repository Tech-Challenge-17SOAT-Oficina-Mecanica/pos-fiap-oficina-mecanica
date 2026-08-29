package ordemservico

import (
	"errors"
	"time"
)

const StatusEntregue = "ENTREGUE"

var ErrOSNaoFinalizada = errors.New("ordem de servico nao esta finalizada")
var ErrOSJaEntregue = errors.New("veiculo ja foi entregue")
var ErrValorFinalIndisponivel = errors.New("valor final da ordem de servico indisponivel")

func ValidarEntrega(status string) error {
	if status == StatusEntregue {
		return ErrOSJaEntregue
	}
	if status != StatusFinalizada {
		return ErrOSNaoFinalizada
	}
	return nil
}

type ResultadoEntrega struct {
	OrdemServicoID       string
	Status               string
	ValorFinal           float64
	ResponsavelEntregaID string
	ClienteID            string
	DataEntrega          time.Time
	Observacoes          string
}
