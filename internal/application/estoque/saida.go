package estoque

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrOrdemServicoNaoEncontrada = errors.New("ordem de servico nao encontrada")
	ErrOSForaDeExecucao          = errors.New("ordem de servico nao esta em execucao")
	ErrReservaAtivaNaoEncontrada = errors.New("item sem reserva ativa para a ordem de servico")
	ErrSaldoInsuficiente         = errors.New("saldo fisico ou reservado insuficiente")
)

type ItemSaidaInput struct {
	ItemID     string
	Quantidade float64
}

type RegistrarSaidaInput struct {
	IdempotencyKey       string
	OrdemServicoID       string
	Itens                []ItemSaidaInput
	LiberarSaldoNaoUsado bool
	UsuarioID            string
}

type ResultadoSaida struct {
	Saida        domain.ResultadoSaida
	JaProcessada bool
}

type SaidaRepository interface {
	RegistrarSaida(context.Context, RegistrarSaidaInput, domain.SaidaCadastro) (ResultadoSaida, error)
}

type RegistrarSaida struct{ repository SaidaRepository }

func NewRegistrarSaida(repository SaidaRepository) RegistrarSaida {
	return RegistrarSaida{repository: repository}
}

func (useCase RegistrarSaida) Execute(ctx context.Context, input RegistrarSaidaInput) (ResultadoSaida, error) {
	if !validation.IsUUID(input.IdempotencyKey) {
		return ResultadoSaida{}, domain.ErrIdempotencyKeyObrigatoria
	}
	if !validation.IsUUID(input.OrdemServicoID) {
		return ResultadoSaida{}, ErrOrdemServicoNaoEncontrada
	}
	itens := make([]domain.ItemSaida, 0, len(input.Itens))
	for _, item := range input.Itens {
		if !validation.IsUUID(item.ItemID) {
			return ResultadoSaida{}, domain.ErrItemIDInvalido
		}
		itens = append(itens, domain.ItemSaida{ItemID: item.ItemID, Quantidade: item.Quantidade})
	}
	cadastro, err := domain.NovaSaidaCadastro(input.OrdemServicoID, itens, input.LiberarSaldoNaoUsado)
	if err != nil {
		return ResultadoSaida{}, err
	}
	return useCase.repository.RegistrarSaida(ctx, input, cadastro)
}
