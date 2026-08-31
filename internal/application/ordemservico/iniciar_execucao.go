package ordemservico

import (
	"context"
	"log"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type IniciarExecucaoInput struct {
	OSID      string
	UsuarioID string
}

type IniciarExecucaoRepository interface {
	IniciarExecucao(context.Context, IniciarExecucaoInput) (domain.ResultadoInicioExecucao, error)
}

type IniciarExecucao struct {
	repository  IniciarExecucaoRepository
	notificador Notificador
	logger      *log.Logger
}

func NewIniciarExecucao(repository IniciarExecucaoRepository, notificador Notificador, logger *log.Logger) IniciarExecucao {
	if logger == nil {
		logger = log.Default()
	}
	return IniciarExecucao{repository: repository, notificador: notificador, logger: logger}
}

func (useCase IniciarExecucao) Execute(ctx context.Context, input IniciarExecucaoInput) (domain.ResultadoInicioExecucao, error) {
	resultado, err := useCase.repository.IniciarExecucao(ctx, input)
	if err != nil {
		return domain.ResultadoInicioExecucao{}, err
	}

	// Fora da transacao: a baixa de estoque e a OS EM_EXECUCAO ja estao gravadas e nao
	// podem ser desfeitas por uma falha de e-mail (RNF-OS-44).
	avisar(ctx, useCase.notificador, resultado.ClienteID,
		notificacaoDominio.EventoExecucaoIniciada, resultado.OrdemServicoID,
		func(erro error) {
			useCase.logger.Printf("notificacao de inicio da execucao da OS %s nao pode ser enfileirada: %v", resultado.OrdemServicoID, erro)
		})

	return resultado, nil
}
