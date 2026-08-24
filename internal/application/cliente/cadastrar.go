package cliente

import (
	"context"
	"errors"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

var (
	ErrClienteDuplicado     = errors.New("cliente já cadastrado com o CPF/CNPJ informado")
	ErrClienteNaoEncontrado = errors.New("cliente não encontrado")
	ErrVersaoDivergente     = errors.New("If-Match divergente")
	ErrClienteComOSAberta   = errors.New("cliente possui Ordem de Serviço em aberto")
	ErrClienteJaInativo     = errors.New("cliente já está inativo")
	ErrClienteJaAtivo       = errors.New("cliente já está ativo")
)

type Repository interface {
	ExisteAtivoPorDocumento(context.Context, string) (bool, error)
	ExisteAtivoPorDocumentoExcetoID(context.Context, string, string) (bool, error)
	BuscarPorID(context.Context, string) (cliente.Cliente, error)
	BuscarPorIDIncluindoInativo(context.Context, string) (cliente.Cliente, error)
	BuscarPorDocumento(context.Context, string) (cliente.Cliente, error)
	BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error)
	Salvar(context.Context, cliente.Cliente) (cliente.Cliente, error)
	Atualizar(context.Context, cliente.Cliente, int) (cliente.Cliente, error)
	Inativar(context.Context, InativarRepositoryInput) (Inativacao, error)
	Reativar(context.Context, ReativarRepositoryInput) (Reativacao, error)
}

type OrdemServicoAberta struct {
	ID     string `json:"ordemServicoId"`
	Status string `json:"status"`
}

type OSAbertaError struct {
	Ordens []OrdemServicoAberta
}

func (err OSAbertaError) Error() string { return ErrClienteComOSAberta.Error() }

func (err OSAbertaError) Is(target error) bool { return target == ErrClienteComOSAberta }

type InativarRepositoryInput struct {
	ClienteID    string
	InativadoPor string
	Motivo       string
}

type Inativacao struct {
	Cliente            cliente.Cliente
	VeiculosInativados []cliente.VeiculoInativado
	DocumentoLiberado  bool
}

type ReativarRepositoryInput struct {
	ClienteID    string
	ReativadoPor string
}

type Reativacao struct {
	Cliente            cliente.Cliente
	ReativadoEm        time.Time
	VeiculosReativados int
}

type Cadastrar struct {
	repository Repository
}

func NewCadastrar(repository Repository) Cadastrar {
	return Cadastrar{repository: repository}
}

func (useCase Cadastrar) Execute(ctx context.Context, input cliente.NovoClienteInput) (cliente.Cliente, error) {
	novo, err := cliente.Novo(input)
	if err != nil {
		return cliente.Cliente{}, err
	}
	exists, err := useCase.repository.ExisteAtivoPorDocumento(ctx, novo.Documento)
	if err != nil {
		return cliente.Cliente{}, err
	}
	if exists {
		return cliente.Cliente{}, ErrClienteDuplicado
	}
	return useCase.repository.Salvar(ctx, novo)
}
