package orcamento

import "time"

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
