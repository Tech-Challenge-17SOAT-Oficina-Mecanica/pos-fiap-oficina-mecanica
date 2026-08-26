package seguranca

// Escopos oficiais da API. Sao a fonte unica: o cadastro de mecanico valida contra esta
// lista e as rotas exigem estas mesmas constantes, entao um escopo escrito errado quebra
// na compilacao em vez de virar um 403 silencioso em producao.
const (
	EscopoMecanicosEscrever = "mecanicos:escrever"

	EscopoClientesLer      = "clientes:ler"
	EscopoClientesEscrever = "clientes:escrever"

	EscopoVeiculosLer      = "veiculos:ler"
	EscopoVeiculosEscrever = "veiculos:escrever"

	EscopoOSLer      = "os:ler"
	EscopoOSEscrever = "os:escrever"

	EscopoOrcamentosLer      = "orcamentos:ler"
	EscopoOrcamentosEscrever = "orcamentos:escrever"
	EscopoOrcamentosDecidir  = "orcamentos:decidir"

	EscopoServicosLer      = "servicos:ler"
	EscopoServicosEscrever = "servicos:escrever"

	EscopoEstoqueLer        = "estoque:ler"
	EscopoEstoqueEscrever   = "estoque:escrever"
	EscopoEstoqueMovimentar = "estoque:movimentar"

	EscopoComprasLer      = "compras:ler"
	EscopoComprasEscrever = "compras:escrever"
)

// EscoposOficiais lista todos os escopos aceitos pela API.
var EscoposOficiais = []string{
	EscopoMecanicosEscrever,
	EscopoClientesLer,
	EscopoClientesEscrever,
	EscopoVeiculosLer,
	EscopoVeiculosEscrever,
	EscopoOSLer,
	EscopoOSEscrever,
	EscopoOrcamentosLer,
	EscopoOrcamentosEscrever,
	EscopoOrcamentosDecidir,
	EscopoServicosLer,
	EscopoServicosEscrever,
	EscopoEstoqueLer,
	EscopoEstoqueEscrever,
	EscopoEstoqueMovimentar,
	EscopoComprasLer,
	EscopoComprasEscrever,
}

// EscopoValido diz se o escopo pertence a lista oficial.
func EscopoValido(escopo string) bool {
	_, encontrado := conjuntoOficial[escopo]
	return encontrado
}

var conjuntoOficial = func() map[string]struct{} {
	conjunto := make(map[string]struct{}, len(EscoposOficiais))
	for _, escopo := range EscoposOficiais {
		conjunto[escopo] = struct{}{}
	}
	return conjunto
}()
