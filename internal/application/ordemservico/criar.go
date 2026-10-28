package ordemservico

import (
	"context"
	"errors"
	"strings"

	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrClienteIDObrigatorio       = errors.New("clienteId é obrigatório")
	ErrVeiculoIDObrigatorio       = errors.New("veiculoId é obrigatório")
	ErrClienteIDInvalido          = errors.New("clienteId inválido")
	ErrVeiculoIDInvalido          = errors.New("veiculoId inválido")
	ErrClienteNaoEncontrado       = errors.New("cliente não encontrado")
	ErrVeiculoNaoEncontrado       = errors.New("veículo não encontrado")
	ErrVeiculoNaoVinculadoCliente = errors.New("veículo não está vinculado ao cliente informado")
	ErrServicoIDInvalido          = errors.New("servicoId invalido")
)

// CriarInput abre a OS. Servicos e Pecas sao opcionais: quando vem algum, a OS ja nasce com o
// orcamento PRINCIPAL contendo esses itens, tudo na mesma transacao.
type CriarInput struct {
	ClienteID string
	VeiculoID string
	Servicos  []domain.ServicoCadastro
	Pecas     []ItemInput
	UsuarioID string
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
	if !validation.IsUUID(input.ClienteID) {
		return domain.OrdemDeServico{}, ErrClienteIDInvalido
	}
	if !validation.IsUUID(input.VeiculoID) {
		return domain.OrdemDeServico{}, ErrVeiculoIDInvalido
	}
	if len(input.Servicos) > 0 {
		servicos, err := domain.NovosServicosCadastro(input.Servicos)
		if err != nil {
			return domain.OrdemDeServico{}, err
		}
		for _, servico := range servicos {
			if !validation.IsUUID(servico.ServicoID) {
				return domain.OrdemDeServico{}, ErrServicoIDInvalido
			}
		}
		input.Servicos = servicos
	}
	if err := validarItens(domainEstoque.TipoPeca, input.Pecas); err != nil {
		return domain.OrdemDeServico{}, err
	}
	return useCase.repository.Criar(ctx, input)
}
