package ordemservico

import (
	"context"
	"errors"

	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domainOrcamento "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	domainOS "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var (
	ErrOSNaoEncontrada        = errors.New("ordem de servico nao encontrada")
	ErrItemNaoEncontrado      = errors.New("item de estoque nao encontrado")
	ErrItemInativo            = errors.New("item de estoque inativo")
	ErrItemSemValor           = errors.New("item de estoque sem valor unitario vigente")
	ErrOrcamentoNaoEncontrado = errors.New("orcamento nao encontrado")
	ErrOrcamentoAprovado      = errors.New("o orcamento principal ja foi aprovado")
	ErrItemRepetido           = errors.New("item repetido na requisicao")
	ErrStatusNaoPermiteItens  = domainOS.ErrStatusNaoPermiteItens
)

type ItemInput struct {
	ItemID     string
	Quantidade float64
}

type RegistrarInput struct {
	OSID      string
	Tipo      string
	Itens     []ItemInput
	UsuarioID string
}

type Repository interface {
	RegistrarItens(context.Context, RegistrarInput) (domainOrcamento.Resultado, error)
	ConsultarOrcamentos(context.Context, string) ([]domainOrcamento.Resultado, error)
}

type RegistrarItens struct{ repository Repository }

func NewRegistrarItens(repository Repository) RegistrarItens {
	return RegistrarItens{repository: repository}
}

func (useCase RegistrarItens) Execute(ctx context.Context, input RegistrarInput) (domainOrcamento.Resultado, error) {
	if len(input.Itens) == 0 {
		return domainOrcamento.Resultado{}, errors.New("itens e obrigatorio")
	}
	seen := make(map[string]struct{}, len(input.Itens))
	for _, item := range input.Itens {
		if _, exists := seen[item.ItemID]; exists {
			return domainOrcamento.Resultado{}, ErrItemRepetido
		}
		seen[item.ItemID] = struct{}{}
		if err := domainEstoque.QuantidadeValida(input.Tipo, item.Quantidade); err != nil {
			return domainOrcamento.Resultado{}, err
		}
	}
	return useCase.repository.RegistrarItens(ctx, input)
}

func (useCase RegistrarItens) Consultar(ctx context.Context, osID string) ([]domainOrcamento.Resultado, error) {
	return useCase.repository.ConsultarOrcamentos(ctx, osID)
}
