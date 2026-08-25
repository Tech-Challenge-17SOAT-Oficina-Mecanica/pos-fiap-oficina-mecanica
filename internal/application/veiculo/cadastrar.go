package veiculo

import (
	"context"
	"errors"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
	"time"
)

var (
	ErrClienteNaoEncontrado       = errors.New("cliente não encontrado")
	ErrClienteInativo             = errors.New("cliente inativo")
	ErrPlacaDuplicada             = errors.New("placa já cadastrada")
	ErrVeiculoJaInativo           = errors.New("veículo já está inativo")
	ErrVeiculoJaAtivo             = errors.New("veículo já está ativo")
	ErrClienteProprietarioInativo = errors.New("cliente proprietário está inativo")
)

type Repository interface {
	CadastrarParaCliente(context.Context, string, domain.Cadastro) (domain.Veiculo, error)
	ConsultarPorPlaca(context.Context, string, bool) (domain.Veiculo, error)
	Atualizar(context.Context, string, int, domain.Cadastro) (domain.Veiculo, error)
	BuscarPorIDIncluindoInativo(context.Context, string) (domain.Veiculo, error)
	BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error)
	Inativar(context.Context, InativarRepositoryInput) (Inativacao, error)
	ExisteAtivoPorPlacaExcetoID(context.Context, string, string) (bool, error)
	ClienteAtivo(context.Context, string) (bool, error)
	Reativar(context.Context, string, string) (Reativacao, error)
}

type OrdemServicoAberta struct {
	OrdemServicoID string `json:"ordemServicoId"`
	Status         string `json:"status"`
}

type OSAbertaError struct{ Ordens []OrdemServicoAberta }

func (err OSAbertaError) Error() string { return "veículo possui ordem de serviço em aberto" }

type InativarRepositoryInput struct {
	VeiculoID    string
	InativadoPor string
	Motivo       string
}

type Inativacao struct {
	Veiculo domain.Veiculo
}

type Reativacao struct {
	Veiculo      domain.Veiculo
	ReativadoEm  time.Time
	ReativadoPor string
}
type Cadastrar struct{ repository Repository }

func NewCadastrar(repository Repository) Cadastrar { return Cadastrar{repository} }
func (useCase Cadastrar) Execute(ctx context.Context, clienteID string, cadastro domain.Cadastro) (domain.Veiculo, error) {
	return useCase.repository.CadastrarParaCliente(ctx, clienteID, cadastro)
}
