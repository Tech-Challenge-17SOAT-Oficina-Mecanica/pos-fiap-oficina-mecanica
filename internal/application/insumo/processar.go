package insumo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const OperacaoSolicitarCompraReservarInsumos = "SOLICITAR_COMPRA_RESERVAR_INSUMOS"

var (
	ErrIdempotencyKeyObrigatoria = errors.New("Idempotency-Key e obrigatorio")
	ErrItemObrigatorio           = errors.New("informe ao menos um insumo")
	ErrItemRepetido              = errors.New("itemId repetido")
	ErrQuantidadeProcessamento   = errors.New("quantidade deve ser maior que zero e ter ate 3 casas decimais")
	ErrFornecedorIdentificador   = errors.New("fornecedorId deve ser um UUID valido")
	ErrFornecedorNaoEncontrado   = errors.New("fornecedor inexistente")
	ErrFornecedorInativo         = errors.New("fornecedor inativo")
	ErrOrdemServicoNaoEncontrada = errors.New("ordem de servico inexistente")
	ErrOrdemServicoInvalida      = errors.New("ordem de servico sem orcamento aprovado")
	ErrItemIdentificador         = errors.New("itemId deve ser um UUID valido")
	ErrItemProcessamentoInvalido = errors.New("insumo inativo, peca ou fora da OS/orcamento aprovado")
	ErrProcessamentoDuplicado    = errors.New("insumo ja possui reserva ativa ou compra aberta para esta OS")
	ErrIdempotencyKeyEmUso       = errors.New("Idempotency-Key ja utilizada com outra requisicao")
)

type ItemProcessamento struct {
	ItemID     string      `json:"itemId"`
	Quantidade json.Number `json:"quantidade"`
}

type SolicitacaoCompraReserva struct {
	IdempotencyKey string
	HashRequisicao string
	OrdemServicoID string
	FornecedorID   string
	Itens          []ItemProcessamento
}

type ItemReservado struct {
	ItemID              string      `json:"itemId"`
	Quantidade          json.Number `json:"quantidade"`
	SaldoDisponivelApos json.Number `json:"saldoDisponivelApos"`
}

type ItemCompraSolicitada struct {
	ItemID       string      `json:"itemId"`
	Quantidade   json.Number `json:"quantidade"`
	ValorParcial json.Number `json:"valorParcial"`
}

type FornecedorProcessamento struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

type ResultadoCompraReserva struct {
	OrdemServicoID          string                  `json:"ordemServicoId"`
	StatusOrdemServico      string                  `json:"statusOrdemServico"`
	InsumosReservados       []ItemReservado         `json:"insumosReservados"`
	InsumosCompraSolicitada []ItemCompraSolicitada  `json:"insumosCompraSolicitada"`
	Fornecedor              FornecedorProcessamento `json:"fornecedor"`
	ValorTotalCompraParcial json.Number             `json:"valorTotalCompraParcial"`
	Reprocessado            bool                    `json:"-"`
}

type ProcessarInsumosRepository interface {
	SolicitarCompraEReservar(ctx context.Context, solicitacao SolicitacaoCompraReserva) (ResultadoCompraReserva, error)
}

type SolicitarCompraEReservarInsumos struct {
	repository ProcessarInsumosRepository
}

func NewSolicitarCompraEReservarInsumos(repository ProcessarInsumosRepository) SolicitarCompraEReservarInsumos {
	return SolicitarCompraEReservarInsumos{repository: repository}
}

func (useCase SolicitarCompraEReservarInsumos) Execute(ctx context.Context, solicitacao SolicitacaoCompraReserva) (ResultadoCompraReserva, error) {
	solicitacao.IdempotencyKey = strings.TrimSpace(solicitacao.IdempotencyKey)
	solicitacao.OrdemServicoID = strings.TrimSpace(solicitacao.OrdemServicoID)
	solicitacao.FornecedorID = strings.TrimSpace(solicitacao.FornecedorID)

	if !validation.IsUUID(solicitacao.IdempotencyKey) {
		return ResultadoCompraReserva{}, ErrIdempotencyKeyObrigatoria
	}
	if !validation.IsUUID(solicitacao.OrdemServicoID) {
		return ResultadoCompraReserva{}, ErrIdentificadorInvalido
	}
	if !validation.IsUUID(solicitacao.FornecedorID) {
		return ResultadoCompraReserva{}, ErrFornecedorIdentificador
	}
	if len(solicitacao.Itens) == 0 {
		return ResultadoCompraReserva{}, ErrItemObrigatorio
	}

	vistos := make(map[string]struct{}, len(solicitacao.Itens))
	for indice := range solicitacao.Itens {
		item := &solicitacao.Itens[indice]
		item.ItemID = strings.TrimSpace(item.ItemID)
		if !validation.IsUUID(item.ItemID) {
			return ResultadoCompraReserva{}, ErrItemIdentificador
		}
		if !QuantidadeValida(item.Quantidade.String()) {
			return ResultadoCompraReserva{}, ErrQuantidadeProcessamento
		}
		if _, existe := vistos[item.ItemID]; existe {
			return ResultadoCompraReserva{}, ErrItemRepetido
		}
		vistos[item.ItemID] = struct{}{}
	}

	return useCase.repository.SolicitarCompraEReservar(ctx, solicitacao)
}
