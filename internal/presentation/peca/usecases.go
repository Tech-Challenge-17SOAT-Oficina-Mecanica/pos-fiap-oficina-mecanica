package peca

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type AtualizarPecaUseCase interface {
	Execute(ctx context.Context, id string, version int, atualizacao domain.Atualizacao, usuarioID string) (domain.Peca, error)
}

type CadastrarPecaUseCase interface {
	Execute(ctx context.Context, cadastro domain.Cadastro) (domain.Peca, error)
}

type ConsultarPecasUseCase interface {
	Execute(ctx context.Context, filtros application.Filtros, limite, deslocamento int) (application.Resultado, error)
	BuscarPorID(ctx context.Context, id string) (domain.Peca, error)
}

type DesativarPecaUseCase interface {
	Execute(ctx context.Context, id, usuarioID string) (domain.Peca, error)
}

type SolicitarCompraEReservarPecasUseCase interface {
	Execute(ctx context.Context, solicitacao application.SolicitacaoCompraReserva) (application.ResultadoCompraReserva, error)
}
