package orcamento

import "math"

// DadosEstimativa reune as entradas do calculo de prazo. Cada uma pode faltar: numa
// oficina nova nao ha OS finalizada para tirar media, e um item nunca comprado nao tem
// prazo conhecido. Nesses casos a parcela vale zero em vez de impedir o calculo — o
// numero fica otimista, e isso e preferivel a nao entregar estimativa nenhuma.
type DadosEstimativa struct {
	// PrazoItensDias e o maior prazo entre os itens indisponiveis. Os pedidos correm em
	// paralelo, entao o que atrasa a OS e o item mais demorado, nao a soma deles.
	PrazoItensDias int
	// TempoServicosDias vem da media de execucao das OS ja finalizadas.
	TempoServicosDias int
	// OSNaFrente e a quantidade de ordens abertas criadas antes desta.
	OSNaFrente int
	// CapacidadeDiaria e quantas OS a oficina atende por dia. Zero significa nao
	// configurada, e entao a fila nao entra na conta.
	CapacidadeDiaria int
}

// DiasFila converte a posicao na fila em dias, conforme a capacidade da oficina.
// Sem capacidade configurada a fila nao e considerada.
func (dados DadosEstimativa) DiasFila() int {
	if dados.CapacidadeDiaria <= 0 || dados.OSNaFrente <= 0 {
		return 0
	}
	return int(math.Ceil(float64(dados.OSNaFrente) / float64(dados.CapacidadeDiaria)))
}

// EstimativaPrincipal soma prazo dos itens, tempo dos servicos e dias de fila.
func EstimativaPrincipal(dados DadosEstimativa) int {
	return naoNegativo(dados.PrazoItensDias) + naoNegativo(dados.TempoServicosDias) + dados.DiasFila()
}

// EstimativaComplementar parte da estimativa que o principal ja tem e soma apenas o que
// o complementar acrescenta. A fila nao entra de novo: a OS ja esta na fila uma vez.
func EstimativaComplementar(estimativaPrincipalDias int, dados DadosEstimativa) int {
	return naoNegativo(estimativaPrincipalDias) + naoNegativo(dados.PrazoItensDias) + naoNegativo(dados.TempoServicosDias)
}

func naoNegativo(valor int) int {
	if valor < 0 {
		return 0
	}
	return valor
}
