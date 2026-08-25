package ordemservico

import (
	"context"
	"errors"
	"regexp"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var (
	ErrClienteIDObrigatorio       = errors.New("clienteId é obrigatório")
	ErrVeiculoIDObrigatorio       = errors.New("veiculoId é obrigatório")
	ErrClienteIDInvalido          = errors.New("clienteId inválido")
	ErrVeiculoIDInvalido          = errors.New("veiculoId inválido")
	ErrClienteNaoEncontrado       = errors.New("cliente não encontrado")
	ErrVeiculoNaoEncontrado       = errors.New("veículo não encontrado")
	ErrVeiculoNaoVinculadoCliente = errors.New("veículo não está vinculado ao cliente informado")
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type CriarInput struct {
	ClienteID string
	VeiculoID string
}

type Repository interface {
	Criar(context.Context, CriarInput) (domain.OrdemDeServico, error)
}

type Criar struct{ repository Repository }

func NewCriar(repository Repository) Criar { return Criar{repository: repository} }

func (useCase Criar) Execute(ctx context.Context, input CriarInput) (domain.OrdemDeServico, error) {
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	input.VeiculoID = strings.TrimSpace(input.VeiculoID)
	if input.ClienteID == "" {
		return domain.OrdemDeServico{}, ErrClienteIDObrigatorio
	}
	if input.VeiculoID == "" {
		return domain.OrdemDeServico{}, ErrVeiculoIDObrigatorio
	}
	if !uuidPattern.MatchString(input.ClienteID) {
		return domain.OrdemDeServico{}, ErrClienteIDInvalido
	}
	if !uuidPattern.MatchString(input.VeiculoID) {
		return domain.OrdemDeServico{}, ErrVeiculoIDInvalido
	}
	return useCase.repository.Criar(ctx, input)
}
