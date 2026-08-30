package estoque

import (
	"errors"
	"time"
)

const ReservaConsumida = "CONSUMIDA"

var ErrQuantidadeMaiorQueReserva = errors.New("quantidade baixada maior que a reservada")

type ItemSaida struct {
	ItemID     string
	Quantidade float64
}

type SaidaCadastro struct {
	OrdemServicoID       string
	Itens                []ItemSaida
	LiberarSaldoNaoUsado bool
}

func NovaSaidaCadastro(ordemServicoID string, itens []ItemSaida, liberarSaldoNaoUsado bool) (SaidaCadastro, error) {
	if len(itens) == 0 {
		return SaidaCadastro{}, ErrItensObrigatorios
	}
	if len(itens) > 200 {
		return SaidaCadastro{}, ErrItensExcedemLimite
	}
	vistos := make(map[string]struct{}, len(itens))
	for _, item := range itens {
		if _, existe := vistos[item.ItemID]; existe {
			return SaidaCadastro{}, ErrItemRepetido
		}
		vistos[item.ItemID] = struct{}{}
		if item.Quantidade <= 0 {
			return SaidaCadastro{}, errors.New("quantidade deve ser maior que zero")
		}
	}
	return SaidaCadastro{OrdemServicoID: ordemServicoID, Itens: itens, LiberarSaldoNaoUsado: liberarSaldoNaoUsado}, nil
}

type ItemSaidaResultado struct {
	ItemID                   string  `json:"itemId"`
	Codigo                   string  `json:"codigo"`
	Tipo                     string  `json:"tipo"`
	UnidadeMedida            string  `json:"unidadeMedida,omitempty"`
	QuantidadeBaixada        float64 `json:"quantidadeBaixada"`
	QuantidadeReservadaAntes float64 `json:"quantidadeReservadaAntes"`
	QuantidadeLiberada       float64 `json:"quantidadeLiberada"`
	SaldoFisicoAtual         float64 `json:"saldoFisicoAtual"`
	SaldoReservadoAtual      float64 `json:"saldoReservadoAtual"`
	CustoUnitario            float64 `json:"custoUnitario"`
	CustoTotal               float64 `json:"custoTotal"`
}

type ResultadoSaida struct {
	SaidaID         string               `json:"saidaId"`
	OrdemServicoID  string               `json:"ordemServicoId"`
	RegistradoEm    time.Time            `json:"registradoEm"`
	RegistradoPor   string               `json:"registradoPor,omitempty"`
	Itens           []ItemSaidaResultado `json:"itens"`
	CustoTotalSaida float64              `json:"custoTotalSaida"`
}
