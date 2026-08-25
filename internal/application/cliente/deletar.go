package cliente

import (
	"context"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type InativarInput struct {
	ClienteID string
	UsuarioID string
	Motivo    string
}

type ReativarInput struct {
	ClienteID string
	UsuarioID string
}

type Inativar struct {
	repository Repository
}

func NewInativar(repository Repository) Inativar {
	return Inativar{repository: repository}
}

func (useCase Inativar) Execute(ctx context.Context, input InativarInput) (Inativacao, error) {
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	if input.ClienteID == "" {
		return Inativacao{}, domain.ErrClienteIDObrigatorio
	}
	motivo, err := domain.MotivoParaInativacao(input.Motivo)
	if err != nil {
		return Inativacao{}, err
	}
	cliente, err := useCase.repository.BuscarPorIDIncluindoInativo(ctx, input.ClienteID)
	if err != nil {
		return Inativacao{}, err
	}
	if !cliente.Ativo {
		return Inativacao{}, ErrClienteJaInativo
	}
	osAbertas, err := useCase.repository.BuscarOSAbertas(ctx, input.ClienteID)
	if err != nil {
		return Inativacao{}, err
	}
	if len(osAbertas) > 0 {
		return Inativacao{}, OSAbertaError{Ordens: osAbertas}
	}
	return useCase.repository.Inativar(ctx, InativarRepositoryInput{ClienteID: input.ClienteID, InativadoPor: input.UsuarioID, Motivo: motivo})
}

type Reativar struct {
	repository Repository
}

func NewReativar(repository Repository) Reativar {
	return Reativar{repository: repository}
}

func (useCase Reativar) Execute(ctx context.Context, input ReativarInput) (Reativacao, error) {
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	if input.ClienteID == "" {
		return Reativacao{}, domain.ErrClienteIDObrigatorio
	}
	cliente, err := useCase.repository.BuscarPorIDIncluindoInativo(ctx, input.ClienteID)
	if err != nil {
		return Reativacao{}, err
	}
	if cliente.Ativo {
		return Reativacao{}, ErrClienteJaAtivo
	}
	exists, err := useCase.repository.ExisteAtivoPorDocumentoExcetoID(ctx, cliente.Documento, cliente.ID)
	if err != nil {
		return Reativacao{}, err
	}
	if exists {
		return Reativacao{}, ErrClienteDuplicado
	}
	return useCase.repository.Reativar(ctx, ReativarRepositoryInput{ClienteID: input.ClienteID, ReativadoPor: input.UsuarioID})
}
