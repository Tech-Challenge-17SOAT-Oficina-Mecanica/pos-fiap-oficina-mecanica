package peca

import (
	"context"
	"strings"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type ErroSaldoReservado struct {
	OrdensServico []string
}

func (erro ErroSaldoReservado) Error() string {
	return "peca possui saldo reservado nas ordens de servico: " + strings.Join(erro.OrdensServico, ", ")
}

type ErroEmOrcamento struct{}

func (ErroEmOrcamento) Error() string {
	return "peca esta em orcamento aguardando decisao do cliente"
}

type DesativarRepository interface {
	BuscarPorID(ctx context.Context, id string) (peca.Peca, error)
	OrdensComReservaAtiva(ctx context.Context, itemID string) ([]string, error)
	EmOrcamentoCriado(ctx context.Context, itemID string) (bool, error)
	Desativar(ctx context.Context, item peca.Peca) error
}

type DesativarPeca struct {
	repository DesativarRepository
	agora      func() time.Time
}

func NewDesativarPeca(repository DesativarRepository) DesativarPeca {
	return DesativarPeca{repository: repository, agora: time.Now}
}

func (useCase DesativarPeca) Execute(ctx context.Context, id, usuarioID string) (peca.Peca, error) {
	id = strings.TrimSpace(id)
	if !padraoUUID.MatchString(id) {
		return peca.Peca{}, ErrIdentificadorInvalido
	}

	encontrada, err := useCase.repository.BuscarPorID(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}

	desativada, err := encontrada.Desativar(usuarioID, useCase.agora())
	if err != nil {
		return peca.Peca{}, err
	}

	ordens, err := useCase.repository.OrdensComReservaAtiva(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}
	if len(ordens) > 0 {
		return peca.Peca{}, ErroSaldoReservado{OrdensServico: ordens}
	}

	emOrcamento, err := useCase.repository.EmOrcamentoCriado(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}
	if emOrcamento {
		return peca.Peca{}, ErroEmOrcamento{}
	}

	if err := useCase.repository.Desativar(ctx, desativada); err != nil {
		return peca.Peca{}, err
	}
	return desativada, nil
}
