package insumo

import (
	"context"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

type CadastrarInsumoUseCase interface {
	Execute(ctx context.Context, cadastro domain.Cadastro) (domain.Insumo, error)
}

type ConsultarInsumosUseCase interface {
	Execute(ctx context.Context, filtros application.FiltrosConsulta, limite, deslocamento int) (application.ResultadoConsulta, error)
	BuscarPorID(ctx context.Context, id string) (domain.Insumo, error)
}

type DesativarInsumoUseCase interface {
	Execute(ctx context.Context, insumoID, usuarioID string) (domain.Insumo, error)
}

type SolicitarCompraEReservarInsumosUseCase interface {
	Execute(ctx context.Context, solicitacao application.SolicitacaoCompraReserva) (application.ResultadoCompraReserva, error)
}
