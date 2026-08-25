package orcamento

import (
	"errors"
	"math/big"
)

const (
	TipoPrincipal    = "PRINCIPAL"
	TipoComplementar = "COMPLEMENTAR"
	StatusCriado     = "CRIADO"
)

var (
	ErrStatusInvalido           = errors.New("orçamento deve estar em CRIADO")
	ErrTipoInvalido             = errors.New("tipo de orçamento inválido")
	ErrVinculoPrincipalInvalido = errors.New("vínculo com orçamento principal inválido")
	ErrSemItens                 = errors.New("orçamento deve possuir ao menos um item")
	ErrItemInvalido             = errors.New("item possui quantidade ou valor unitário inválido")
	ErrPrazoIndisponivel        = errors.New("prazo de entrega do material indisponível")
	ErrTempoServicoIndisponivel = errors.New("tempo do serviço indisponível")
	ErrConfiguracaoInvalida     = errors.New("configuração da oficina inválida")
)

type Item struct {
	Tipo                string
	Quantidade          string
	ValorUnitario       string
	TempoServicoMinutos int
	MaterialDisponivel  bool
	PrazoEntregaDias    *int
}

type Calculo struct {
	Tipo                    string
	Status                  string
	OrcamentoOriginalID     string
	OrcamentoPrincipalID    string
	EstimativaPrincipalDias *int
	Itens                   []Item
	CapacidadeDiariaOS      int
	MinutosProdutivosDia    int
	OrdensNaFila            int
}

func (calculo Calculo) EstimativaEntregaDias() (int, error) {
	if calculo.Status != StatusCriado {
		return 0, ErrStatusInvalido
	}
	if calculo.Tipo != TipoPrincipal && calculo.Tipo != TipoComplementar {
		return 0, ErrTipoInvalido
	}
	if calculo.Tipo == TipoPrincipal && calculo.OrcamentoOriginalID != "" {
		return 0, ErrVinculoPrincipalInvalido
	}
	if calculo.Tipo == TipoComplementar && (calculo.OrcamentoOriginalID == "" || calculo.OrcamentoOriginalID != calculo.OrcamentoPrincipalID || calculo.EstimativaPrincipalDias == nil) {
		return 0, ErrVinculoPrincipalInvalido
	}
	if len(calculo.Itens) == 0 {
		return 0, ErrSemItens
	}
	if calculo.CapacidadeDiariaOS <= 0 || calculo.MinutosProdutivosDia <= 0 {
		return 0, ErrConfiguracaoInvalida
	}
	minutosServicos, prazoMateriais := 0, 0
	for _, item := range calculo.Itens {
		if !decimalPositivo(item.Quantidade) || !decimalNaoNegativo(item.ValorUnitario) {
			return 0, ErrItemInvalido
		}
		switch item.Tipo {
		case "SERVICO":
			if item.TempoServicoMinutos <= 0 {
				return 0, ErrTempoServicoIndisponivel
			}
			minutosServicos += item.TempoServicoMinutos
		case "PECA", "INSUMO":
			if !item.MaterialDisponivel {
				if item.PrazoEntregaDias == nil || *item.PrazoEntregaDias < 0 {
					return 0, ErrPrazoIndisponivel
				}
				if *item.PrazoEntregaDias > prazoMateriais {
					prazoMateriais = *item.PrazoEntregaDias
				}
			}
		default:
			return 0, ErrItemInvalido
		}
	}
	diasServicos := arredondarParaCima(minutosServicos, calculo.MinutosProdutivosDia)
	if calculo.Tipo == TipoComplementar {
		return *calculo.EstimativaPrincipalDias + prazoMateriais + diasServicos, nil
	}
	diasFila := arredondarParaCima(calculo.OrdensNaFila, calculo.CapacidadeDiariaOS)
	return prazoMateriais + diasServicos + diasFila, nil
}

func arredondarParaCima(valor, divisor int) int {
	if valor == 0 {
		return 0
	}
	return (valor + divisor - 1) / divisor
}

func decimalPositivo(value string) bool {
	number, ok := new(big.Rat).SetString(value)
	return ok && number.Sign() > 0
}

func decimalNaoNegativo(value string) bool {
	number, ok := new(big.Rat).SetString(value)
	return ok && number.Sign() >= 0
}
