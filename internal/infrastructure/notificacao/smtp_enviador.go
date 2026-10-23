package notificacao

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

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

func (enviador SMTPEnviador) Enviar(ctx context.Context, aviso notificacao.Notificacao) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	host, _, err := net.SplitHostPort(enviador.endereco)
	if err != nil {
		return err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", enviador.endereco)
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	if err = conn.SetDeadline(deadline); err != nil {
		return err
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if enviador.autenticar != nil {
		if err = client.Auth(enviador.autenticar); err != nil {
			return err
		}
	}
	if err = client.Mail(enviador.enderecoRemetente); err != nil {
		return err
	}
	if err = client.Rcpt(aviso.Destinatario); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(mensagemRFC822(enviador.remetente, aviso)); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	// DATA ja foi aceito: falha ao fechar a sessao nao exige reenviar a mensagem.
	_ = client.Quit()
	return nil
}

// mensagemRFC822 monta o e-mail. O assunto vai codificado em UTF-8 porque os avisos tem
// acento, e cabecalho SMTP e ASCII por padrao.
//
// Com HTML presente o corpo vai em multipart/alternative: o cliente de e-mail escolhe a
// melhor versao que consegue exibir, e quem le em texto puro ainda recebe algo legivel.
func mensagemRFC822(remetente string, aviso notificacao.Notificacao) []byte {
	var mensagem strings.Builder
	fmt.Fprintf(&mensagem, "From: %s\r\n", remetente)
	fmt.Fprintf(&mensagem, "To: %s\r\n", aviso.Destinatario)
	fmt.Fprintf(&mensagem, "Subject: =?UTF-8?B?%s?=\r\n", base64UTF8(aviso.Assunto))
	mensagem.WriteString("MIME-Version: 1.0\r\n")

	if aviso.ConteudoHTML == "" {
		mensagem.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		mensagem.WriteString(aviso.Conteudo)
		mensagem.WriteString("\r\n")
		return []byte(mensagem.String())
	}

	const fronteira = "oficina-mecanica-alternative"
	fmt.Fprintf(&mensagem, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", fronteira)

	// A versao em texto vem primeiro: o RFC 2046 manda listar da menos para a mais rica.
	fmt.Fprintf(&mensagem, "--%s\r\n", fronteira)
	mensagem.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	mensagem.WriteString(aviso.Conteudo)
	mensagem.WriteString("\r\n")

	fmt.Fprintf(&mensagem, "--%s\r\n", fronteira)
	mensagem.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	mensagem.WriteString(aviso.ConteudoHTML)
	mensagem.WriteString("\r\n")

	fmt.Fprintf(&mensagem, "--%s--\r\n", fronteira)
	return []byte(mensagem.String())
}

func base64UTF8(valor string) string {
	return base64.StdEncoding.EncodeToString([]byte(valor))
}
