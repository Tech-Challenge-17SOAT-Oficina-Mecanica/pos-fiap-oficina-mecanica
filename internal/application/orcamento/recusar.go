package orcamento

import (
	"context"
	"errors"
	"log"
	"strings"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

var (
	ErrOrcamentoComplementarSemPrincipal  = errors.New("orcamento complementar sem orcamento principal vinculado")
	ErrMotivoInvalido                     = errors.New("motivo deve ter no maximo 500 caracteres")
	ErrOrdemServicoNaoAguardandoAprovacao = errors.New("ordem de servico nao esta aguardando aprovacao")
)

type RecusarInput struct {
	OrcamentoID    string
	ClienteID      string
	OrdemServicoID string
	Motivo         string
}

type RecusarRepository interface {
	Recusar(context.Context, RecusarInput) (domain.Decisao, error)
}

type Recusar struct {
	repository  RecusarRepository
	notificador Notificador
	logger      *log.Logger
}

func NewRecusar(repository RecusarRepository, notificador Notificador, logger *log.Logger) Recusar {
	if logger == nil {
		logger = log.Default()
	}
	return Recusar{repository: repository, notificador: notificador, logger: logger}
}

func (useCase Recusar) Execute(ctx context.Context, input RecusarInput) (domain.Decisao, error) {
	input.Motivo = strings.TrimSpace(input.Motivo)
	if len(input.Motivo) > 500 {
		return domain.Decisao{}, ErrMotivoInvalido
	}
	resultado, err := useCase.repository.Recusar(ctx, input)
	if err != nil {
		return domain.Decisao{}, err
	}

	// So a recusa do PRINCIPAL encerra o atendimento. A do COMPLEMENTAR devolve a OS para
	// AGUARDANDO_EXECUCAO, mas isso e consequencia direta da decisao que o cliente acabou
	// de tomar: nao ha novidade a comunicar.
	if resultado.StatusOrdemServico == "CANCELADA" {
		avisar(ctx, useCase.notificador, resultado.ClienteID,
			notificacaoDominio.EventoServicoCancelado, resultado.OrdemServicoID,
			func(erro error) {
				useCase.logger.Printf("notificacao do cancelamento da OS %s nao pode ser enfileirada: %v", resultado.OrdemServicoID, erro)
			})
	}

	return resultado, nil
}
