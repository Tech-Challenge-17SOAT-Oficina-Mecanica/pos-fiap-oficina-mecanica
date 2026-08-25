package insumo

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

var (
	ErrDesativacaoInvalida = errors.New("dados de desativação inválidos")
	ErrInsumoNaoEncontrado = errors.New("insumo não encontrado")
	ErrInsumoEmUso         = errors.New("insumo possui reserva ou orçamento aguardando aprovação")
)

type InsumoEmUsoError struct{ OrdensServico []string }

func (err InsumoEmUsoError) Error() string { return ErrInsumoEmUso.Error() }
func (err InsumoEmUsoError) Unwrap() error { return ErrInsumoEmUso }

type DesativarRepository interface {
	Desativar(context.Context, string, string) (domain.Insumo, error)
}

type DesativarInsumo struct{ repository DesativarRepository }

func NewDesativarInsumo(repository DesativarRepository) DesativarInsumo {
	return DesativarInsumo{repository: repository}
}

func (useCase DesativarInsumo) Execute(ctx context.Context, insumoID, usuarioID string) (domain.Insumo, error) {
	insumoID, usuarioID = strings.TrimSpace(insumoID), strings.TrimSpace(usuarioID)
	if insumoID == "" || usuarioID == "" {
		return domain.Insumo{}, ErrDesativacaoInvalida
	}
	return useCase.repository.Desativar(ctx, insumoID, usuarioID)
}
