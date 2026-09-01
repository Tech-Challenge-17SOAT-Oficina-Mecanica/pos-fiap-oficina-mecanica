package notificacao

import (
	"strings"
	"testing"

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
