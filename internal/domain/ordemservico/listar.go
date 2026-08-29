package ordemservico

// StatusValidosListagem sao os status aceitos no filtro de GET /ordens-servico. Inclui
// AGUARDANDO_RECURSOS, que o refinamento original nao listou mas que e um status real do
// sistema (introduzido pelo fluxo de estoque).
var StatusValidosListagem = []string{
	"RECEBIDA", "EM_DIAGNOSTICO", "AGUARDANDO_APROVACAO", "AGUARDANDO_RECURSOS",
	"AGUARDANDO_EXECUCAO", "EM_EXECUCAO", "FINALIZADA", "ENTREGUE", "CANCELADA",
}

func StatusListagemValido(status string) bool {
	for _, valido := range StatusValidosListagem {
		if valido == status {
			return true
		}
	}
	return false
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
