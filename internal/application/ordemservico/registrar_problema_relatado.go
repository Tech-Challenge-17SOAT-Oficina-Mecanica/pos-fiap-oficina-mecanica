package ordemservico

import (
	"context"
	"errors"
	"log"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

var (
	ErrOrdemServicoForaDeRecebida   = errors.New("ordem de serviço não está com status RECEBIDA")
	ErrProblemaRelatadoJaRegistrado = errors.New("problema relatado já registrado")
)

type RegistrarProblemaRelatadoInput struct {
	OrdemServicoID string
	Descricao      string
	Observacoes    string
}

type ProblemaRelatadoRepository interface {
	RegistrarProblemaRelatado(context.Context, string, domain.ProblemaRelatado) (domain.OrdemDeServico, error)
}

type RegistrarProblemaRelatado struct {
	repository  ProblemaRelatadoRepository
	notificador Notificador
	logger      *log.Logger
}

func NewRegistrarProblemaRelatado(repository ProblemaRelatadoRepository, notificador Notificador, logger *log.Logger) RegistrarProblemaRelatado {
	if logger == nil {
		logger = log.Default()
	}
	return RegistrarProblemaRelatado{repository: repository, notificador: notificador, logger: logger}
}

func (useCase RegistrarProblemaRelatado) Execute(ctx context.Context, input RegistrarProblemaRelatadoInput) (domain.OrdemDeServico, error) {
	problema, err := domain.NovoProblemaRelatado(input.Descricao, input.Observacoes)
	if err != nil {
		return domain.OrdemDeServico{}, err
	}

	resultado, err := useCase.repository.RegistrarProblemaRelatado(ctx, input.OrdemServicoID, problema)
	if err != nil {
		return domain.OrdemDeServico{}, err
	}

	// Fora da transacao: a OS ja esta EM_DIAGNOSTICO e continua assim mesmo que o aviso
	// falhe (RNF-OS-44). O resultado nao entra na resposta porque o retorno aqui e a
	// propria entidade, e o envio do aviso nao e atributo dela.
	avisar(ctx, useCase.notificador, resultado.ClienteID,
		notificacaoDominio.EventoDiagnosticoIniciado, resultado.ID,
		func(erro error) {
			useCase.logger.Printf("notificacao de inicio do diagnostico da OS %s nao pode ser enfileirada: %v", resultado.ID, erro)
		})

	return resultado, nil
}
