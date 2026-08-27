package peca

import (
	"context"
	"errors"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const OperacaoSolicitarCompraReservarPecas = "SOLICITAR_COMPRA_RESERVAR_PECAS"

var (
	ErrIdempotencyKeyObrigatoria = errors.New("Idempotency-Key e obrigatorio")
	ErrItemObrigatorio           = errors.New("informe ao menos uma peca")
	ErrItemRepetido              = errors.New("itemId repetido")
	ErrQuantidadeProcessamento   = errors.New("quantidade deve ser inteira e maior que zero")
	ErrFornecedorIdentificador   = errors.New("fornecedorId deve ser um UUID valido")
	ErrFornecedorNaoEncontrado   = errors.New("fornecedor inexistente")
	ErrFornecedorInativo         = errors.New("fornecedor inativo")
	ErrOrdemServicoNaoEncontrada = errors.New("ordem de servico inexistente")
	ErrOrdemServicoInvalida      = errors.New("ordem de servico sem orcamento aprovado")
	ErrItemIdentificador         = errors.New("itemId deve ser um UUID valido")
	ErrItemNaoEncontrado         = errors.New("peca inexistente")
	ErrItemProcessamentoInvalido = errors.New("peca inativa, insumo ou fora da OS/orcamento aprovado")
	ErrProcessamentoDuplicado    = errors.New("peca ja possui reserva ativa ou compra aberta para esta OS")
	ErrIdempotencyKeyEmUso       = errors.New("Idempotency-Key ja utilizada com outra requisicao")
)

type ItemProcessamento struct {
	ItemID     string `json:"itemId"`
	Quantidade int64  `json:"quantidade"`
}

type SolicitacaoCompraReserva struct {
	IdempotencyKey string
	HashRequisicao string
	OrdemServicoID string
	FornecedorID   string
	Itens          []ItemProcessamento
}

type ItemReservado struct {
	ItemID              string `json:"itemId"`
	Quantidade          int64  `json:"quantidade"`
	SaldoDisponivelApos int64  `json:"saldoDisponivelApos"`
}

type ItemCompraSolicitada struct {
	ItemID       string  `json:"itemId"`
	Quantidade   int64   `json:"quantidade"`
	ValorParcial float64 `json:"valorParcial"`
}

type FornecedorProcessamento struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

type ResultadoCompraReserva struct {
	OrdemServicoID          string                  `json:"ordemServicoId"`
	StatusOrdemServico      string                  `json:"statusOrdemServico"`
	PecasReservadas         []ItemReservado         `json:"pecasReservadas"`
	PecasCompraSolicitada   []ItemCompraSolicitada  `json:"pecasCompraSolicitada"`
	Fornecedor              FornecedorProcessamento `json:"fornecedor"`
	ValorTotalCompraParcial float64                 `json:"valorTotalCompraParcial"`
	Reprocessado            bool                    `json:"-"`
}

type ProcessarPecasRepository interface {
	SolicitarCompraEReservar(ctx context.Context, solicitacao SolicitacaoCompraReserva) (ResultadoCompraReserva, error)
}

type SolicitarCompraEReservarPecas struct {
	repository ProcessarPecasRepository
}

func NewSolicitarCompraEReservarPecas(repository ProcessarPecasRepository) SolicitarCompraEReservarPecas {
	return SolicitarCompraEReservarPecas{repository: repository}
}

func (useCase SolicitarCompraEReservarPecas) Execute(ctx context.Context, solicitacao SolicitacaoCompraReserva) (ResultadoCompraReserva, error) {
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
		if item.Quantidade <= 0 {
			return ResultadoCompraReserva{}, ErrQuantidadeProcessamento
		}
		if _, existe := vistos[item.ItemID]; existe {
			return ResultadoCompraReserva{}, ErrItemRepetido
		}
		vistos[item.ItemID] = struct{}{}
	}

	return useCase.repository.SolicitarCompraEReservar(ctx, solicitacao)
}
