package orcamento

import (
	"context"
	"errors"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

// ErrOrcamentoNaoEncontrado ja existe em consultar.go e serve aos dois fluxos.
var ErrIdentificadorInvalido = errors.New("identificador deve ser um UUID valido")

// OrcamentoDaOS e um orcamento irmao, usado apenas para somar o total geral da OS.
type OrcamentoDaOS struct {
	ID     string
	Tipo   string
	Status string
	Itens  []orcamento.Item
}

type CalcularRepository interface {
	// BuscarParaCalculo devolve o orcamento alvo com seus itens.
	BuscarParaCalculo(ctx context.Context, orcamentoID string) (orcamento.Orcamento, string, error)
	// OrcamentosDaOrdem devolve todos os orcamentos da OS, com itens, para o total geral.
	OrcamentosDaOrdem(ctx context.Context, ordemServicoID string) ([]OrcamentoDaOS, error)
	// SalvarItens grava os valores recalculados e marca a atualizacao do orcamento.
	SalvarItens(ctx context.Context, orcamentoID string, itens []orcamento.Item) error
}

type ResultadoCalculo struct {
	OrcamentoID     string
	OrdemServicoID  string
	ValorTotal      float64
	ValorTotalGeral float64
}

type Calcular struct {
	repository CalcularRepository
}

func NewCalcular(repository CalcularRepository) Calcular {
	return Calcular{repository: repository}
}

func (useCase Calcular) Execute(ctx context.Context, orcamentoID string) (ResultadoCalculo, error) {
	if !validation.IsUUID(orcamentoID) {
		return ResultadoCalculo{}, ErrIdentificadorInvalido
	}

	alvo, ordemServicoID, err := useCase.repository.BuscarParaCalculo(ctx, orcamentoID)
	if err != nil {
		return ResultadoCalculo{}, err
	}

	// Toda a leitura vem antes da escrita, de proposito: se qualquer consulta falhar,
	// nada foi gravado ainda. A escrita e a ultima operacao, e e atomica (RNF-ORC-01).
	irmaos, err := useCase.repository.OrcamentosDaOrdem(ctx, ordemServicoID)
	if err != nil {
		return ResultadoCalculo{}, err
	}

	vinculos := make([]orcamento.Vinculo, 0, len(irmaos))
	for _, irmao := range irmaos {
		vinculos = append(vinculos, orcamento.Vinculo{ID: irmao.ID, Tipo: irmao.Tipo})
	}
	if err := alvo.ValidarParaCalculo(vinculos); err != nil {
		return ResultadoCalculo{}, err
	}

	itens, total, err := orcamento.Recalcular(alvo.Itens)
	if err != nil {
		return ResultadoCalculo{}, err
	}

	// O total geral soma os orcamentos validos da OS. O alvo entra com os valores que
	// acabaram de ser recalculados, nao com os que estavam gravados antes.
	totalGeral := total
	for _, irmao := range irmaos {
		if irmao.ID == orcamentoID {
			continue
		}
		if !(orcamento.Orcamento{Status: irmao.Status}).EntraNoTotalGeral() {
			continue
		}
		for _, item := range irmao.Itens {
			totalGeral += orcamento.TotalItem(item)
		}
	}

	// Unica escrita do fluxo, depois de tudo validado e calculado.
	if err := useCase.repository.SalvarItens(ctx, orcamentoID, itens); err != nil {
		return ResultadoCalculo{}, err
	}

	return ResultadoCalculo{
		OrcamentoID:     orcamentoID,
		OrdemServicoID:  ordemServicoID,
		ValorTotal:      total,
		ValorTotalGeral: arredondar(totalGeral),
	}, nil
}

func arredondar(valor float64) float64 {
	return float64(int64(valor*100+0.5)) / 100
}
