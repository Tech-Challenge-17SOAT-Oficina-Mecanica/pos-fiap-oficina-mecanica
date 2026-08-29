package ordemservico

import "time"

// ItemFila representa uma OS apta para iniciar a execucao.
type ItemFila struct {
	OrdemServicoID        string
	Placa                 string
	Marca                 string
	Modelo                string
	Status                string
	MecanicoResponsavelID *string
	DataEntradaFila       time.Time
}
