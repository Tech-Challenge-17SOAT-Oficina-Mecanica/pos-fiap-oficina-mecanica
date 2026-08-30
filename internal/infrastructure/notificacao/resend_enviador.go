package notificacao

import (
	"context"
	"fmt"

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
}

func NewResendEnviador(apiKey, remetente string) ResendEnviador {
	return ResendEnviador{cliente: resend.NewClient(apiKey), remetente: remetente}
}

func (enviador ResendEnviador) Enviar(ctx context.Context, aviso notificacao.Notificacao) error {
	enviado, err := enviador.cliente.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    enviador.remetente,
		To:      []string{aviso.Destinatario},
		Subject: aviso.Assunto,
		Text:    aviso.Conteudo,
	})
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	if enviado == nil || enviado.Id == "" {
		return fmt.Errorf("resend: envio aceito sem identificador de retorno")
	}
	return nil
}
