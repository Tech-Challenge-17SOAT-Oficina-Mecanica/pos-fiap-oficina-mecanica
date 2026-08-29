package ordemservico

import (
	"context"
	"log"
	"strings"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type EntregarInput struct {
	OSID        string
	ClienteID   string
	Observacoes string
	UsuarioID   string
}

type EntregarRepository interface {
	Entregar(context.Context, EntregarInput) (domain.ResultadoEntrega, error)
}

type Entregar struct {
	repository  EntregarRepository
	notificador Notificador
	logger      *log.Logger
}

func NewEntregar(repository EntregarRepository, notificador Notificador, logger *log.Logger) Entregar {
	if logger == nil {
		logger = log.Default()
	}
	return Entregar{repository: repository, notificador: notificador, logger: logger}
}

func (useCase Entregar) Execute(ctx context.Context, input EntregarInput) (domain.ResultadoEntrega, error) {
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	input.Observacoes = strings.TrimSpace(input.Observacoes)
	if input.ClienteID != "" && !validation.IsUUID(input.ClienteID) {
		return domain.ResultadoEntrega{}, ErrClienteIDInvalido
	}

	resultado, err := useCase.repository.Entregar(ctx, input)
	if err != nil {
		return domain.ResultadoEntrega{}, err
	}

	// Confirmacao da entrega, tambem fora da transacao: a OS ja esta ENTREGUE.
	avisar(ctx, useCase.notificador, resultado.ClienteID,
		notificacaoDominio.EventoVeiculoEntregue, resultado.OrdemServicoID,
		func(erro error) {
			useCase.logger.Printf("notificacao de entrega da OS %s nao pode ser enfileirada: %v", resultado.OrdemServicoID, erro)
		})

	return resultado, nil
}
