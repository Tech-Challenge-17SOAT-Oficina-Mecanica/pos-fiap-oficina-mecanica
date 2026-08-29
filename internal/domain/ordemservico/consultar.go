package ordemservico

import (
	"encoding/json"
	"time"
)

// ClienteResumo e VeiculoResumo sustentam GET /ordens-servico/{osId}.
type ClienteResumo struct {
	ID        string
	Nome      string
	Documento string
}

type VeiculoResumo struct {
	ID     string
	Placa  string
	Marca  string
	Modelo string
	Ano    int
}

type ProblemaConsulta struct {
	ID             string
	Descricao      string
	OrcamentoID    string
	IdentificadoEm time.Time
}

type ItemOrcamentoConsulta struct {
	Tipo          string
	Descricao     string
	Quantidade    float64
	ValorUnitario float64
	ValorTotal    float64
}

type OrcamentoConsulta struct {
	ID                  string
	Tipo                string
	OrcamentoOriginalID string
	Itens               []ItemOrcamentoConsulta
	ValorTotal          float64
	DataGeracao         time.Time
}

// EventoConsulta e uma linha da trilha de auditoria (`auditoria_ordem_servico`), exposta como
// `eventos`. Nao ha colunas separadas de statusAnterior/statusNovo no schema: quando presentes,
// esses dados vivem dentro de `Dados`.
type EventoConsulta struct {
	ID           string
	Agregado     string
	AgregadoID   string
	TipoEvento   string
	Dados        json.RawMessage
	Metadados    json.RawMessage
	OcorridoEm   time.Time
	RegistradoEm time.Time
}

// ConsultaDetalhada e o retorno de GET /ordens-servico/{osId}.
type ConsultaDetalhada struct {
	OrdemServicoID     string
	StatusOrdemServico string
	Cliente            ClienteResumo
	Veiculo            VeiculoResumo
	Problemas          []ProblemaConsulta
	Orcamentos         []OrcamentoConsulta
	ValorTotalGeral    float64
	Eventos            []EventoConsulta
}
