package orcamento

import "time"

const (
	TipoPrincipal    = "PRINCIPAL"
	TipoComplementar = "COMPLEMENTAR"
	StatusCriado     = "CRIADO"
	StatusAprovado   = "APROVADO"
	StatusRecusado   = "RECUSADO"
)

type Item struct {
	Tipo          string  `json:"tipo"`
	Descricao     string  `json:"descricao"`
	Quantidade    float64 `json:"quantidade"`
	ValorUnitario float64 `json:"valorUnitario"`
	ValorTotal    float64 `json:"valorTotal"`
}

type Orcamento struct {
	ID                    string    `json:"orcamentoId"`
	Tipo                  string    `json:"tipoOrcamento"`
	OrcamentoOriginalID   string    `json:"orcamentoOriginalId,omitempty"`
	Status                string    `json:"statusOrcamento"`
	Itens                 []Item    `json:"itens"`
	ValorTotal            float64   `json:"valorTotal"`
	EstimativaEntregaDias *int      `json:"estimativaEntregaDias"`
	DataGeracao           time.Time `json:"dataGeracao"`
}

type Cliente struct {
	ID        string `json:"clienteId"`
	Documento string `json:"documento"`
}

type Consulta struct {
	Cliente            Cliente     `json:"cliente"`
	OrdemServicoID     string      `json:"ordemServicoId"`
	StatusOrdemServico string      `json:"statusOrdemServico"`
	Orcamentos         []Orcamento `json:"orcamentos"`
	ValorTotalGeral    float64     `json:"valorTotalGeral"`
}
