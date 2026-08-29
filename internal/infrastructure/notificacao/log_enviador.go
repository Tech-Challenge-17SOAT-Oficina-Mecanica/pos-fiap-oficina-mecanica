package notificacao

import (
	"context"
	"log"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// LogEnviador registra o aviso em log em vez de enviar. Permite exercitar a fila inteira
// sem depender de servidor SMTP; o adaptador real implementa a mesma interface.
type LogEnviador struct {
	logger *log.Logger
}

func NewLogEnviador(logger *log.Logger) LogEnviador {
	return LogEnviador{logger: logger}
}

// Enviar nao registra o conteudo da mensagem, so o suficiente para rastrear o envio —
// o corpo tem dados do cliente e nao deve ir para o log.
func (enviador LogEnviador) Enviar(_ context.Context, aviso notificacao.Notificacao) error {
	enviador.logger.Printf("notificacao %s canal=%s evento=%s origem=%s/%s assunto=%q",
		aviso.ID, aviso.Canal, aviso.TipoEvento, aviso.Origem.Agregado, aviso.Origem.ID, aviso.Assunto)
	return nil
}
