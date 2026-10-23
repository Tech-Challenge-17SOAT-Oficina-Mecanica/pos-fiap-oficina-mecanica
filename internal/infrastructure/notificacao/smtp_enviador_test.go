package notificacao

import (
	"context"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

func TestApenasEndereco(t *testing.T) {
	if got := apenasEndereco("Oficina <oficina@example.com>"); got != "oficina@example.com" {
		t.Fatalf("endereco=%q", got)
	}
	if got := apenasEndereco("valor invalido"); got != "valor invalido" {
		t.Fatalf("endereco invalido deve ser preservado: %q", got)
	}
}

func TestMensagemRFC822(t *testing.T) {
	aviso := notificacao.Notificacao{
		Destinatario: "cliente@example.com",
		Assunto:      "Orçamento pronto",
		Conteudo:     "Texto simples",
	}

	texto := string(mensagemRFC822("Oficina <oficina@example.com>", aviso))
	if !strings.Contains(texto, "Content-Type: text/plain") || !strings.Contains(texto, "Texto simples") {
		t.Fatalf("mensagem texto invalida: %s", texto)
	}

	aviso.ConteudoHTML = "<strong>HTML</strong>"
	html := string(mensagemRFC822("Oficina <oficina@example.com>", aviso))
	for _, trecho := range []string{"multipart/alternative", "Texto simples", "<strong>HTML</strong>", "--oficina-mecanica-alternative--"} {
		if !strings.Contains(html, trecho) {
			t.Fatalf("mensagem html sem %q: %s", trecho, html)
		}
	}
}

func TestNewResendEnviador(t *testing.T) {
	enviador := NewResendEnviador("token", "Oficina <oficina@example.com>", nil)
	if enviador.cliente == nil || enviador.logger == nil || enviador.remetente == "" {
		t.Fatalf("enviador invalido: %+v", enviador)
	}
}

func TestSMTPCancelaServidorSemResposta(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var b [1]byte
		_, _ = conn.Read(b[:])
	}()
	sender := SMTPEnviador{endereco: listener.Addr().String()}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err = sender.Enviar(ctx, notificacao.Notificacao{}); err == nil {
		t.Fatal("esperado cancelamento")
	}
	if time.Since(start) > time.Second {
		t.Fatal("SMTP ignorou o prazo")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("conexao nao encerrada")
	}
}

func TestSMTPEntregaMensagemSemServicoExterno(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	received := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			received <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		peer := textproto.NewConn(conn)
		defer peer.Close()
		if err = peer.PrintfLine("220 teste SMTP"); err != nil {
			received <- err
			return
		}
		for _, step := range []struct{ command, response string }{
			{"EHLO ", "250 teste"}, {"MAIL FROM:<oficina@example.com>", "250 ok"},
			{"RCPT TO:<cliente@example.com>", "250 ok"}, {"DATA", "354 envie"},
		} {
			line, e := peer.ReadLine()
			if e != nil {
				received <- e
				return
			}
			if !strings.HasPrefix(line, step.command) {
				received <- fmt.Errorf("comando inesperado: %s", line)
				return
			}
			if e = peer.PrintfLine("%s", step.response); e != nil {
				received <- e
				return
			}
		}
		body, e := peer.ReadDotBytes()
		if e != nil {
			received <- e
			return
		}
		if !strings.Contains(string(body), "Mensagem teste") {
			received <- fmt.Errorf("conteudo ausente")
			return
		}
		_ = peer.PrintfLine("250 aceita")
		line, e := peer.ReadLine()
		if e != nil {
			received <- e
			return
		}
		if line != "QUIT" {
			received <- fmt.Errorf("esperado QUIT, recebido %s", line)
			return
		}
		_ = peer.PrintfLine("221 fim")
		received <- nil
	}()
	sender := SMTPEnviador{endereco: listener.Addr().String(), remetente: "Oficina <oficina@example.com>", enderecoRemetente: "oficina@example.com"}
	err = sender.Enviar(context.Background(), notificacao.Notificacao{Destinatario: "cliente@example.com", Assunto: "Teste", Conteudo: "Mensagem teste"})
	if err != nil {
		t.Fatal(err)
	}
	if err = <-received; err != nil {
		t.Fatal(err)
	}
}
