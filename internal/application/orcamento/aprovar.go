package orcamento

import (
	"context"
	"errors"
	"log"
	"strings"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrIdentificadorInvalido       = errors.New("orcamentoId deve ser um UUID valido")
	ErrOrcamentoJaDecidido         = errors.New("orcamento ja aprovado ou recusado")
	ErrOrdemServicoStatusInvalido  = errors.New("ordem de servico fora de AGUARDANDO_APROVACAO")
	ErrOrcamentoComplementarSemPai = errors.New("orcamento complementar sem vinculo valido com o principal")
)

type AprovarInput struct {
	OrcamentoID    string
	ClienteID      string
	UsuarioID      string
	OrdemServicoID string
}

type AprovarRepository interface {
	Aprovar(context.Context, AprovarInput) (domain.Aprovacao, error)
}

type Aprovar struct {
	repository  AprovarRepository
	notificador Notificador
	logger      *log.Logger
}

func NewAprovar(repository AprovarRepository, notificador Notificador, logger *log.Logger) Aprovar {
	if logger == nil {
		logger = log.Default()
	}
	return Aprovar{repository: repository, notificador: notificador, logger: logger}
}

func (useCase Aprovar) Execute(ctx context.Context, input AprovarInput) (domain.Aprovacao, error) {
	input.OrcamentoID = strings.TrimSpace(input.OrcamentoID)
	input.ClienteID = strings.TrimSpace(input.ClienteID)
	input.UsuarioID = strings.TrimSpace(input.UsuarioID)
	input.OrdemServicoID = strings.TrimSpace(input.OrdemServicoID)

	if !validation.IsUUID(input.OrcamentoID) {
		return domain.Aprovacao{}, ErrIdentificadorInvalido
	}
	if input.ClienteID != "" && !validation.IsUUID(input.ClienteID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	if input.ClienteID == "" && !validation.IsUUID(input.UsuarioID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	if input.OrdemServicoID != "" && !validation.IsUUID(input.OrdemServicoID) {
		return domain.Aprovacao{}, ErrAcessoNegado
	}
	resultado, err := useCase.repository.Aprovar(ctx, input)
	if err != nil {
		return domain.Aprovacao{}, err
	}

	// Fora da transacao: a aprovacao ja esta gravada e continua valendo mesmo que o aviso
	// falhe (RNF-OS-44). O evento depende de para onde a OS foi: aprovar nem sempre leva
	// para a fila de execucao, e o cliente precisa saber qual dos dois aconteceu.
	evento := notificacaoDominio.EventoOrcamentoAprovado
	if resultado.StatusOrdemServico == "AGUARDANDO_RECURSOS" {
		evento = notificacaoDominio.EventoAguardandoRecursos
	}
	avisar(ctx, useCase.notificador, resultado.ClienteID, evento, resultado.OrdemServicoID,
		func(erro error) {
			useCase.logger.Printf("notificacao da aprovacao do orcamento %s nao pode ser enfileirada: %v", resultado.OrcamentoID, erro)
		})

	return resultado, nil
}
