package ordemservico

import (
	"context"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// Notificador e a porta que a Ordem de Servico usa para avisar o cliente. Fica aqui, e
// nao no pacote de notificacao, para que este contexto dependa de uma interface propria
// e nao do pacote inteiro.
type Notificador interface {
	Execute(ctx context.Context, pedido notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error)
}

// avisar enfileira a notificacao e devolve se deu certo. Nunca propaga o erro: o aviso e
// consequencia da operacao, e falhar nele nao pode desfazer o que ja foi persistido
// (RNF-OS-44). A notificacao fica na fila e pode ser reenviada (A7).
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
