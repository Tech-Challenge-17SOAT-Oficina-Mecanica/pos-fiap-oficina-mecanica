package ordemservico

import "time"

const StatusRecebida = "RECEBIDA"

type OrdemDeServico struct {
	ID                    string
	ClienteID             string
	VeiculoID             string
	PlacaVeiculo          string
	Status                string
	CriadaEm              time.Time
	ProblemaRelatado      ProblemaRelatado
	DataInicioDiagnostico *time.Time
}
