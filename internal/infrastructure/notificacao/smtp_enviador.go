package notificacao

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// SMTPEnviador entrega por um servidor SMTP. Em desenvolvimento aponta para o Mailpit,
// que captura tudo e mostra numa interface web sem nada sair para a internet.
type SMTPEnviador struct {
	endereco string
	// remetente e a forma completa, com nome de exibicao, usada no cabecalho From.
	remetente string
	// enderecoRemetente e so o e-mail. O envelope SMTP nao aceita nome de exibicao:
	// mandar "Oficina <a@b.com>" ali faz o servidor responder 501.
	enderecoRemetente string
	autenticar        smtp.Auth
}

// NewSMTPEnviador monta o adaptador. Usuario vazio significa servidor sem autenticacao,
// que e o caso do Mailpit local.
func NewSMTPEnviador(host string, porta int, remetente, usuario, senha string) SMTPEnviador {
	var autenticar smtp.Auth
	if usuario != "" {
		autenticar = smtp.PlainAuth("", usuario, senha, host)
	}
	return SMTPEnviador{
		endereco:          fmt.Sprintf("%s:%d", host, porta),
		remetente:         remetente,
		enderecoRemetente: apenasEndereco(remetente),
		autenticar:        autenticar,
	}
}

// apenasEndereco extrai o e-mail de uma forma como "Nome <a@b.com>". Se nao conseguir
// interpretar, devolve o valor original — o servidor dira o que ha de errado.
func apenasEndereco(remetente string) string {
	if endereco, err := mail.ParseAddress(remetente); err == nil {
		return endereco.Address
	}
	return remetente
}

func (enviador SMTPEnviador) Enviar(_ context.Context, aviso notificacao.Notificacao) error {
	return smtp.SendMail(enviador.endereco, enviador.autenticar, enviador.enderecoRemetente,
		[]string{aviso.Destinatario}, mensagemRFC822(enviador.remetente, aviso))
}

// mensagemRFC822 monta o e-mail em texto puro. O assunto vai codificado em UTF-8 porque
// os avisos tem acento, e cabecalho SMTP e ASCII por padrao.
func mensagemRFC822(remetente string, aviso notificacao.Notificacao) []byte {
	var mensagem strings.Builder
	fmt.Fprintf(&mensagem, "From: %s\r\n", remetente)
	fmt.Fprintf(&mensagem, "To: %s\r\n", aviso.Destinatario)
	fmt.Fprintf(&mensagem, "Subject: =?UTF-8?B?%s?=\r\n", base64UTF8(aviso.Assunto))
	mensagem.WriteString("MIME-Version: 1.0\r\n")
	mensagem.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	mensagem.WriteString("\r\n")
	mensagem.WriteString(aviso.Conteudo)
	mensagem.WriteString("\r\n")
	return []byte(mensagem.String())
}

func base64UTF8(valor string) string {
	return base64.StdEncoding.EncodeToString([]byte(valor))
}
