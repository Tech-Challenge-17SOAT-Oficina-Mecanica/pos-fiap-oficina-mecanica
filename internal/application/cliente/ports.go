package cliente

import (
	"context"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

// Cada caso de uso define somente as operacoes de persistencia que consome.
type CadastrarRepository interface {
	ExisteAtivoPorDocumento(context.Context, string) (bool, error)
	Salvar(context.Context, domain.Cliente) (domain.Cliente, error)
}

type ConsultarRepository interface {
	BuscarPorDocumento(context.Context, string) (domain.Cliente, error)
}

type AtualizarRepository interface {
	BuscarPorID(context.Context, string) (domain.Cliente, error)
	ExisteAtivoPorDocumentoExcetoID(context.Context, string, string) (bool, error)
	Atualizar(context.Context, domain.Cliente, int) (domain.Cliente, error)
}

type InativarRepository interface {
	BuscarPorIDIncluindoInativo(context.Context, string) (domain.Cliente, error)
	BuscarOSAbertas(context.Context, string) ([]OrdemServicoAberta, error)
	Inativar(context.Context, InativarRepositoryInput) (Inativacao, error)
}

type ReativarRepository interface {
	BuscarPorIDIncluindoInativo(context.Context, string) (domain.Cliente, error)
	ExisteAtivoPorDocumentoExcetoID(context.Context, string, string) (bool, error)
	Reativar(context.Context, ReativarRepositoryInput) (Reativacao, error)
}
