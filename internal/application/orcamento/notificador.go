package orcamento

import (
	"context"
	"errors"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

var errNotificadorAusente = errors.New("notificador nao configurado")

// Notificador e a porta de aviso ao cliente, declarada aqui para o contexto nao depender
// do pacote de notificacao inteiro.
type Notificador interface {
	Execute(ctx context.Context, pedido notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error)
}

// avisar enfileira o aviso da mudanca de status e devolve se deu certo. Nunca propaga o
// erro: a decisao do cliente ja esta gravada, e falhar no e-mail nao pode desfaze-la
// (RNF-OS-44). A notificacao fica na fila e e reenviada pelo processador.
func avisar(ctx context.Context, notificador Notificador, clienteID, tipoEvento, ordemServicoID string, registrarFalha func(error)) bool {
	if notificador == nil {
		return false
	}

	_, err := notificador.Execute(ctx, notificacaoApplication.Pedido{
		ClienteID:  clienteID,
		TipoEvento: tipoEvento,
		Origem:     notificacaoDominio.Origem{Agregado: "ordem_servico", ID: ordemServicoID},
	})
	if err != nil {
		if registrarFalha != nil {
			registrarFalha(err)
		}
		return false
	}
	return true
}
