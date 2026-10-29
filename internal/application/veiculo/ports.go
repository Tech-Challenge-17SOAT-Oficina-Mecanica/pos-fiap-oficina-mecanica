package veiculo

import (
	"context"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

// Os contratos pertencem aos casos de uso, independentemente do banco escolhido.
type CadastrarRepository interface {
	CadastrarParaCliente(context.Context, string, domain.Cadastro) (domain.Veiculo, error)
}

type ConsultarRepository interface {
	ConsultarPorPlaca(context.Context, string, bool) (domain.Veiculo, error)
}

type AtualizarRepository interface {
	Atualizar(context.Context, string, int, domain.Cadastro) (domain.Veiculo, error)
}

type InativarRepository interface {
	BuscarPorIDIncluindoInativo(context.Context, string) (domain.Veiculo, error)
	BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error)
	Inativar(context.Context, InativarRepositoryInput) (Inativacao, error)
}

type ReativarRepository interface {
	BuscarPorIDIncluindoInativo(context.Context, string) (domain.Veiculo, error)
	ClienteAtivo(context.Context, string) (bool, error)
	ExisteAtivoPorPlacaExcetoID(context.Context, string, string) (bool, error)
	Reativar(context.Context, string, string) (Reativacao, error)
}
