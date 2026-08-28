package estoque

import (
	"errors"
	"strings"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const MovimentacaoEntrada = "ENTRADA"

var ErrDocumentoOrigemObrigatorio = errors.New("documentoOrigem e obrigatorio")
var ErrItensObrigatorios = errors.New("itens e obrigatorio")
var ErrItensExcedemLimite = errors.New("itens excede o limite de 200 linhas")
var ErrItemIDInvalido = errors.New("itemId invalido")
var ErrFornecedorIDInvalido = errors.New("fornecedorId invalido")
var ErrItemRepetido = errors.New("item repetido na requisicao")
var ErrCustoInvalido = errors.New("custoUnitario deve ser maior que zero")
var ErrIdempotencyKeyObrigatoria = errors.New("Idempotency-Key e obrigatorio e deve ser um uuid")

// ItemEntrada e uma linha do recebimento: item, quantidade e custo daquela nota.
type ItemEntrada struct {
	ItemID        string
	Quantidade    float64
	CustoUnitario float64
}

// EntradaCadastro e o pedido de registro de entrada, ja validado nos aspectos tecnicos.
type EntradaCadastro struct {
	DocumentoOrigem      string
	FornecedorID         string
	PedidoCompraID       string
	ConfirmarDivergencia bool
	Itens                []ItemEntrada
}

// NovaEntradaCadastro valida os aspectos tecnicos comuns a peca e insumo; a validacao
// especifica de tipo (inteiro para peca, casas decimais para insumo) depende do item
// carregado do banco e acontece no repositorio.
func NovaEntradaCadastro(documentoOrigem, fornecedorID, pedidoCompraID string, confirmarDivergencia bool, itens []ItemEntrada) (EntradaCadastro, error) {
	documentoOrigem = strings.TrimSpace(documentoOrigem)
	fornecedorID = strings.TrimSpace(fornecedorID)
	if fornecedorID != "" && !validation.IsUUID(fornecedorID) {
		return EntradaCadastro{}, ErrFornecedorIDInvalido
	}
	if documentoOrigem == "" {
		return EntradaCadastro{}, ErrDocumentoOrigemObrigatorio
	}
	if len(itens) == 0 {
		return EntradaCadastro{}, ErrItensObrigatorios
	}
	if len(itens) > 200 {
		return EntradaCadastro{}, ErrItensExcedemLimite
	}
	vistos := make(map[string]struct{}, len(itens))
	for _, item := range itens {
		if _, existe := vistos[item.ItemID]; existe {
			return EntradaCadastro{}, ErrItemRepetido
		}
		vistos[item.ItemID] = struct{}{}
		if item.Quantidade <= 0 {
			return EntradaCadastro{}, errors.New("quantidade deve ser maior que zero")
		}
		if item.CustoUnitario <= 0 {
			return EntradaCadastro{}, ErrCustoInvalido
		}
	}
	return EntradaCadastro{
		DocumentoOrigem: documentoOrigem, FornecedorID: fornecedorID, PedidoCompraID: pedidoCompraID,
		ConfirmarDivergencia: confirmarDivergencia, Itens: itens,
	}, nil
}

// ItemEntradaResultado e o saldo de um item apos a entrada.
type ItemEntradaResultado struct {
	ItemID              string  `json:"itemId"`
	Codigo              string  `json:"codigo"`
	UnidadeMedida       string  `json:"unidadeMedida,omitempty"`
	Quantidade          float64 `json:"quantidade"`
	SaldoFisicoAnterior float64 `json:"saldoFisicoAnterior"`
	SaldoFisicoAtual    float64 `json:"saldoFisicoAtual"`
	SaldoReservado      float64 `json:"saldoReservado"`
	SaldoDisponivel     float64 `json:"saldoDisponivel"`
}

// PedidoCompraResultado resume a situacao do pedido apos o recebimento.
type PedidoCompraResultado struct {
	ID     string `json:"id"`
	Numero string `json:"numero"`
	Status string `json:"status"`
}

// OrdemServicoLiberada informa se uma OS vinculada ao pedido mudou de status.
type OrdemServicoLiberada struct {
	OrdemServicoID string `json:"ordemServicoId"`
	StatusAnterior string `json:"statusAnterior"`
	Status         string `json:"status"`
	ItensPendentes int    `json:"itensPendentes"`
}

// ResultadoEntrada e o retorno de RegistrarEntrada.
type ResultadoEntrada struct {
	EntradaID       string                 `json:"entradaId"`
	DocumentoOrigem string                 `json:"documentoOrigem"`
	RegistradoEm    time.Time              `json:"registradoEm"`
	RegistradoPor   string                 `json:"registradoPor,omitempty"`
	Itens           []ItemEntradaResultado `json:"itens"`
	PedidoCompra    *PedidoCompraResultado `json:"pedidoCompra,omitempty"`
	OrdensServico   []OrdemServicoLiberada `json:"ordensServico,omitempty"`
}
