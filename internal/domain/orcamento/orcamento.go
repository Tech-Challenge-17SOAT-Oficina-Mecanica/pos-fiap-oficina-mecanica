package orcamento

const (
	TipoPrincipal    = "PRINCIPAL"
	TipoComplementar = "COMPLEMENTAR"
	StatusCriado     = "CRIADO"
)

type Item struct {
	ItemID        string  `json:"itemId"`
	Codigo        string  `json:"codigo"`
	Descricao     string  `json:"descricao"`
	Tipo          string  `json:"tipo"`
	Quantidade    float64 `json:"quantidade"`
	ValorUnitario float64 `json:"valorUnitario"`
	ValorItem     float64 `json:"valorItem"`
}

type Resultado struct {
	OrdemServicoID    string  `json:"ordemServicoId"`
	OrcamentoID       string  `json:"orcamentoId"`
	OrcamentoOriginal string  `json:"orcamentoOriginalId,omitempty"`
	TipoOrcamento     string  `json:"tipoOrcamento"`
	StatusOrcamento   string  `json:"statusOrcamento"`
	ItensRegistrados  []Item  `json:"itensRegistrados"`
	ValorOrcamento    float64 `json:"valorOrcamento"`
	ValorTotalGeral   float64 `json:"valorTotalGeral"`
	RegistradoPor     string  `json:"registradoPor,omitempty"`
}
