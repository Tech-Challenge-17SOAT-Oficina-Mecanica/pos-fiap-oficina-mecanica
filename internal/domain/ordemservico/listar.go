package ordemservico

import "slices"

// StatusListagemPorPrioridade e a ordem em que GET /ordens-servico apresenta as OS, da mais
// adiantada no fluxo para a menos: o enunciado da Fase 2 pede Execucao > Aguardando Aprovacao >
// Diagnostico > Recebida. AGUARDANDO_EXECUCAO e AGUARDANDO_RECURSOS ficam onde estao na maquina
// de estados, entre a aprovacao e a execucao. FINALIZADA, ENTREGUE e CANCELADA nao aparecem: a
// OS continua no banco, so sai da listagem. Esta lista tambem define os valores aceitos no
// filtro por status.
var StatusListagemPorPrioridade = []string{
	StatusEmExecucao,
	StatusAguardandoExecucao,
	"AGUARDANDO_RECURSOS",
	"AGUARDANDO_APROVACAO",
	StatusEmDiagnostico,
	StatusRecebida,
}

func StatusListagemValido(status string) bool {
	return slices.Contains(StatusListagemPorPrioridade, status)
}

// FiltrosListagem sao os filtros opcionais de GET /ordens-servico; vazio significa "sem filtro".
type FiltrosListagem struct {
	Status    string
	Documento string
	Placa     string
}

// ItemListagem e uma linha da listagem, com cliente e veiculo resumidos.
type ItemListagem struct {
	OrdemServicoID   string
	Status           string
	ClienteID        string
	ClienteNome      string
	ClienteDocumento string
	VeiculoID        string
	Placa            string
	Marca            string
	Modelo           string
}
