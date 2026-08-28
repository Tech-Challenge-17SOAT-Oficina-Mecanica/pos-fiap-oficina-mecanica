package orcamento

import "time"

const (
	TipoPrincipal    = "PRINCIPAL"
	TipoComplementar = "COMPLEMENTAR"
	StatusCriado     = "CRIADO"
	StatusAprovado   = "APROVADO"
	StatusRecusado   = "RECUSADO"

	// OSStatusAguardandoAprovacao e o unico status de OS em que um orcamento PRINCIPAL pode ser recusado.
	OSStatusAguardandoAprovacao = "AGUARDANDO_APROVACAO"
)

// ItemRegistrado é o item devolvido pelo registro de peças/insumos na OS.
type ItemRegistrado struct {
	ItemID        string  `json:"itemId"`
	Codigo        string  `json:"codigo"`
	Descricao     string  `json:"descricao"`
	Tipo          string  `json:"tipo"`
	Quantidade    float64 `json:"quantidade"`
	ValorUnitario float64 `json:"valorUnitario"`
	ValorItem     float64 `json:"valorItem"`
}

// Resultado é a resposta do registro de peças/insumos: o orçamento afetado e seus itens.
type Resultado struct {
	OrdemServicoID    string           `json:"ordemServicoId"`
	OrcamentoID       string           `json:"orcamentoId"`
	OrcamentoOriginal string           `json:"orcamentoOriginalId,omitempty"`
	TipoOrcamento     string           `json:"tipoOrcamento"`
	StatusOrcamento   string           `json:"statusOrcamento"`
	ItensRegistrados  []ItemRegistrado `json:"itensRegistrados"`
	ValorOrcamento    float64          `json:"valorOrcamento"`
	ValorTotalGeral   float64          `json:"valorTotalGeral"`
	RegistradoPor     string           `json:"registradoPor,omitempty"`
}

// Item, Orcamento, Problema, Cliente e Consulta sustentam GET /ordens-servico/{osId}/orcamento.
type Item struct {
	ID            string
	Tipo          string
	Descricao     string
	Quantidade    float64
	ValorUnitario float64
	ValorTotal    float64
}

type Orcamento struct {
	ID             string
	OriginalID     string
	Tipo           string
	Status         string
	EstimativaDias *int
	DataGeracao    time.Time
	Itens          []Item
	Problemas      []Problema
	ValorTotal     float64
}

type Problema struct {
	ID           string
	Descricao    string
	Observacoes  string
	RegistradoEm time.Time
}

type Cliente struct {
	ID            string
	Nome          string
	Documento     string
	TipoDocumento string
}

type Consulta struct {
	Cliente               Cliente
	OrdemServicoID        string
	StatusOrdemServico    string
	Orcamentos            []Orcamento
	ValorTotalGeral       float64
	EstimativaEntregaDias *int
}

// Decisao é o resultado de aprovar ou recusar um orçamento.
type Decisao struct {
	OrcamentoID         string
	OrdemServicoID      string
	TipoOrcamento       string
	OrcamentoOriginalID string
	StatusOrcamento     string
	StatusOrdemServico  string
	ClienteID           string
	DecididoEm          time.Time
	Motivo              string
}
