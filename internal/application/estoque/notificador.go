package estoque

import (
	"context"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// Notificador e a porta de aviso ao cliente. Fica aqui, e nao no pacote de notificacao,
// para o contexto de estoque depender de uma interface propria e nao do pacote inteiro.
type Notificador interface {
	Execute(ctx context.Context, pedido notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error)
}

// avisar enfileira o aviso e nunca propaga o erro: a entrada de estoque ja esta gravada e
// a OS ja foi liberada, e uma falha de e-mail nao pode desfazer nada disso (RNF-OS-44).
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
