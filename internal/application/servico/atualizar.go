package servico

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

var (
	ErrAtualizacaoInvalida = errors.New("dados de atualização inválidos")
	ErrServicoInativo      = errors.New("serviço inativo")
	ErrVersaoDivergente    = errors.New("serviço alterado por outro usuário")
)

type AtualizacaoRepository interface {
	Atualizar(context.Context, string, domain.AtualizacaoInput, int, string) (domain.Servico, error)
}

type AtualizarServico struct {
	repository AtualizacaoRepository
}

func NewAtualizarServico(repository AtualizacaoRepository) AtualizarServico {
	return AtualizarServico{repository: repository}
}

func (uc AtualizarServico) Execute(ctx context.Context, servicoID string, input domain.AtualizacaoInput, version int, usuarioID string) (domain.Servico, error) {
	servicoID = strings.TrimSpace(servicoID)
	if servicoID == "" || version < 1 {
		return domain.Servico{}, ErrAtualizacaoInvalida
	}
	return uc.repository.Atualizar(ctx, servicoID, input, version, usuarioID)
}
