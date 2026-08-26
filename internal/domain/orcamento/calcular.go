package orcamento

import "errors"

const (
	TipoPrincipal    = "PRINCIPAL"
	TipoComplementar = "COMPLEMENTAR"

	StatusCriado   = "CRIADO"
	StatusAprovado = "APROVADO"
	StatusRecusado = "RECUSADO"
)

var (
	ErrStatusNaoCalculavel      = errors.New("apenas orcamento com status CRIADO pode ser calculado")
	ErrSemItens                 = errors.New("o orcamento nao possui itens para calcular")
	ErrComplementarSemPrincipal = errors.New("orcamento complementar exige um principal vinculado da mesma ordem de servico")
	ErrItemInvalido             = errors.New("todos os itens precisam de quantidade e valor unitario validos")
)

// Calculavel diz se o orcamento pode ser alvo do calculo. Aprovado ou recusado ja teve
// decisao do cliente: recalcular mudaria o valor de algo que ele ja respondeu.
func (orcamento Orcamento) Calculavel() bool {
	return orcamento.Status == StatusCriado
}

// EntraNoTotalGeral diz se os itens deste orcamento somam no total da OS. Recusado fica
// de fora (RF-ORC-06); criado e aprovado entram (RF-ORC-05).
func (orcamento Orcamento) EntraNoTotalGeral() bool {
	return orcamento.Status == StatusCriado || orcamento.Status == StatusAprovado
}

// TotalItem recalcula o valor do item. E a unica formula do calculo: o resto e decidir
// quais orcamentos entram na soma.
func TotalItem(item Item) float64 {
	return arredondar(item.Quantidade * item.ValorUnitario)
}

// Recalcular devolve os itens com o valor total atualizado e a soma do orcamento.
func Recalcular(itens []Item) ([]Item, float64, error) {
	if len(itens) == 0 {
		return nil, 0, ErrSemItens
	}

	recalculados := make([]Item, 0, len(itens))
	total := 0.0
	for _, item := range itens {
		if item.Quantidade <= 0 || item.ValorUnitario < 0 {
			return nil, 0, ErrItemInvalido
		}
		item.ValorTotal = TotalItem(item)
		total += item.ValorTotal
		recalculados = append(recalculados, item)
	}
	return recalculados, arredondar(total), nil
}

// ValidarParaCalculo aplica as regras que independem de banco.
func (orcamento Orcamento) ValidarParaCalculo() error {
	if !orcamento.Calculavel() {
		return ErrStatusNaoCalculavel
	}
	if orcamento.Tipo == TipoComplementar && orcamento.OriginalID == "" {
		return ErrComplementarSemPrincipal
	}
	return nil
}

// arredondar mantem duas casas, a precisao do NUMERIC(12,2) das colunas de valor.
func arredondar(valor float64) float64 {
	return float64(int64(valor*100+copiarSinal(0.5, valor))) / 100
}

func copiarSinal(magnitude, sinal float64) float64 {
	if sinal < 0 {
		return -magnitude
	}
	return magnitude
}
