package ordemservico

import (
	"context"
	"log"
	"strings"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type FinalizarInput struct {
	OSID        string
	Observacoes string
	UsuarioID   string
}

type FinalizarRepository interface {
	Finalizar(context.Context, FinalizarInput) (domain.ResultadoFinalizacao, error)
}

type Finalizar struct {
	repository  FinalizarRepository
	notificador Notificador
	logger      *log.Logger
}

func NewFinalizar(repository FinalizarRepository, notificador Notificador, logger *log.Logger) Finalizar {
	if logger == nil {
		logger = log.Default()
	}
	return Finalizar{repository: repository, notificador: notificador, logger: logger}
}

func (useCase Finalizar) Execute(ctx context.Context, input FinalizarInput) (domain.ResultadoFinalizacao, error) {
	input.Observacoes = strings.TrimSpace(input.Observacoes)

	resultado, err := useCase.repository.Finalizar(ctx, input)
	if err != nil {
		return domain.ResultadoFinalizacao{}, err
	}

	// Fora da transacao, de proposito: a OS ja esta FINALIZADA e continua assim mesmo
	// que o aviso falhe (RNF-OS-44). O resultado do enfileiramento vai na resposta,
	// atendendo o RF-OS-88.
	resultado.NotificacaoEnviada = avisar(ctx, useCase.notificador, resultado.ClienteID,
		notificacaoDominio.EventoServicoFinalizado, resultado.OrdemServicoID,
		func(erro error) {
			useCase.logger.Printf("notificacao da OS %s nao pode ser enfileirada: %v", resultado.OrdemServicoID, erro)
		})

	return resultado, nil
}
