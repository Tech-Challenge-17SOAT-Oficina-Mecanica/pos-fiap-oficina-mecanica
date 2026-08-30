package notificacao

import (
	"context"
	"fmt"
	"log"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	"github.com/resend/resend-go/v4"
)

// ResendEnviador entrega pela API do Resend. Diferente do SMTP apontado para o Mailpit,
// este envia de verdade: use com destinatarios reais em mente.
//
// Em modo de teste o Resend so entrega para o e-mail dono da conta, com o remetente
// padrao onboarding@resend.dev. Para qualquer destinatario e preciso verificar um dominio.
type ResendEnviador struct {
	cliente   *resend.Client
	remetente string
	logger    *log.Logger
}

func NewResendEnviador(apiKey, remetente string, logger *log.Logger) ResendEnviador {
	if logger == nil {
		logger = log.Default()
	}
	return ResendEnviador{cliente: resend.NewClient(apiKey), remetente: remetente, logger: logger}
}

func (enviador ResendEnviador) Enviar(ctx context.Context, aviso notificacao.Notificacao) error {
	// Os dois corpos vao juntos quando ha HTML: o Resend entrega multipart e o cliente
	// escolhe. Sem HTML, segue so o texto.
	enviado, err := enviador.cliente.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    enviador.remetente,
		To:      []string{aviso.Destinatario},
		Subject: aviso.Assunto,
		Text:    aviso.Conteudo,
		Html:    aviso.ConteudoHTML,
	})
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	if enviado == nil || enviado.Id == "" {
		return fmt.Errorf("resend: envio aceito sem identificador de retorno")
	}

	// O identificador do provedor e o unico jeito de rastrear a entrega depois: o
	// "aceito" do Resend nao garante que a caixa do destinatario recebeu. Com ele da
	// para procurar o e-mail no painel e ver se caiu em spam, bounce ou foi entregue.
	enviador.logger.Printf("resend aceitou a notificacao %s: id=%s destinatario=%s",
		aviso.ID, enviado.Id, aviso.Destinatario)
	return nil
}
