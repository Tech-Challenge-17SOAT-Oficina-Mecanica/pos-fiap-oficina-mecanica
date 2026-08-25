package servico

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

var (
	ErrServicoJaInativo    = errors.New("serviço já está inativo")
	ErrServicoJaAtivo      = errors.New("serviço já está ativo")
	ErrNomeAtivoDuplicado  = errors.New("já existe serviço ativo com o mesmo nome normalizado")
	ErrSituacaoInvalida    = errors.New("situação do serviço inválida")
)

type SituacaoRepository interface {
	Desativar(context.Context, string, string) (domain.Servico, error)
	Reativar(context.Context, string, string) (domain.Servico, error)
}

type DesativarServico struct {
	repository SituacaoRepository
}

func NewDesativarServico(repository SituacaoRepository) DesativarServico {
	return DesativarServico{repository: repository}
}

func (uc DesativarServico) Execute(ctx context.Context, servicoID, usuarioID string) (domain.Servico, error) {
	servicoID = strings.TrimSpace(servicoID)
	if servicoID == "" {
		return domain.Servico{}, ErrSituacaoInvalida
	}
	return uc.repository.Desativar(ctx, servicoID, usuarioID)
}

type ReativarServico struct {
	repository SituacaoRepository
}

func NewReativarServico(repository SituacaoRepository) ReativarServico {
	return ReativarServico{repository: repository}
}

func (uc ReativarServico) Execute(ctx context.Context, servicoID, usuarioID string) (domain.Servico, error) {
	servicoID = strings.TrimSpace(servicoID)
	if servicoID == "" {
		return domain.Servico{}, ErrSituacaoInvalida
	}
	return uc.repository.Reativar(ctx, servicoID, usuarioID)
}
