package orcamento

import (
	"context"
	"errors"
	"log"
	"time"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var errNotificadorAusente = errors.New("notificador nao configurado")

// OrcamentoParaEnvio reune o que o envio precisa saber, sem carregar o agregado inteiro.
type OrcamentoParaEnvio struct {
	Orcamento      orcamento.Orcamento
	OrdemServicoID string
	ClienteID      string
	StatusOS       string
	Calculado      bool
}

// Notificador e a porta de aviso ao cliente, declarada aqui para o contexto nao depender
// do pacote de notificacao inteiro.
type Notificador interface {
	Execute(ctx context.Context, pedido notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error)
}

type EnviarRepository interface {
	// BuscarParaEnvio traz o orcamento com seus itens e o estado da OS dona.
	BuscarParaEnvio(ctx context.Context, orcamentoID string) (OrcamentoParaEnvio, error)
	// MarcarEnviado poe a OS em AGUARDANDO_APROVACAO e registra o envio.
	MarcarEnviado(ctx context.Context, orcamentoID, ordemServicoID, usuarioID string) (time.Time, error)
}

type ResultadoEnvio struct {
	OrcamentoID        string
	OrdemServicoID     string
	StatusOrdemServico string
	EnviadoEm          time.Time
	NotificacaoEnviada bool
}

// Enviar coloca o orcamento sob decisao do cliente. E a transicao que faltava entre
// calcular e aprovar: sem ela a OS nunca chega a AGUARDANDO_APROVACAO.
type Enviar struct {
	repository  EnviarRepository
	notificador Notificador
	logger      *log.Logger
}

func NewEnviar(repository EnviarRepository, notificador Notificador, logger *log.Logger) Enviar {
	if logger == nil {
		logger = log.Default()
	}
	return Enviar{repository: repository, notificador: notificador, logger: logger}
}

func (useCase Enviar) Execute(ctx context.Context, orcamentoID, usuarioID string) (ResultadoEnvio, error) {
	if !validation.IsUUID(orcamentoID) {
		return ResultadoEnvio{}, ErrIdentificadorInvalido
	}

	dados, err := useCase.repository.BuscarParaEnvio(ctx, orcamentoID)
	if err != nil {
		return ResultadoEnvio{}, err
	}
	if err := dados.Orcamento.ValidarParaEnvio(dados.StatusOS, dados.Calculado); err != nil {
		return ResultadoEnvio{}, err
	}

	enviadoEm, err := useCase.repository.MarcarEnviado(ctx, orcamentoID, dados.OrdemServicoID, usuarioID)
	if err != nil {
		return ResultadoEnvio{}, err
	}

	// Fora da transacao: a OS ja esta AGUARDANDO_APROVACAO e continua assim mesmo que o
	// aviso falhe. Mesma regra do finalizar (RNF-OS-44).
	notificada := true
	if _, err := useCase.notificar(ctx, dados); err != nil {
		useCase.logger.Printf("notificacao do orcamento %s nao pode ser enfileirada: %v", orcamentoID, err)
		notificada = false
	}

	return ResultadoEnvio{
		OrcamentoID:        orcamentoID,
		OrdemServicoID:     dados.OrdemServicoID,
		StatusOrdemServico: orcamento.OSStatusAguardandoAprovacao,
		EnviadoEm:          enviadoEm,
		NotificacaoEnviada: notificada,
	}, nil
}

func (useCase Enviar) notificar(ctx context.Context, dados OrcamentoParaEnvio) (notificacaoDominio.Notificacao, error) {
	if useCase.notificador == nil {
		return notificacaoDominio.Notificacao{}, errNotificadorAusente
	}
	return useCase.notificador.Execute(ctx, notificacaoApplication.Pedido{
		ClienteID:  dados.ClienteID,
		TipoEvento: notificacaoDominio.EventoOrcamentoPronto,
		Origem:     notificacaoDominio.Origem{Agregado: "orcamento", ID: dados.Orcamento.ID},
	})
}
