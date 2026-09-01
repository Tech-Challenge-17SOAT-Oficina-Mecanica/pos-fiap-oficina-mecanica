package notificacao

import (
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

func TestMensagemRFC822(t *testing.T) {
	aviso := notificacao.Notificacao{
		Destinatario: "cliente@example.com",
		Assunto:      "Seu veículo está pronto para retirada",
		Conteudo:     "Olá, Maria! O serviço foi concluído.",
	}

	mensagem := string(mensagemRFC822("Oficina <oficina@example.com>", aviso))

	for _, esperado := range []string{
		"From: Oficina <oficina@example.com>\r\n",
		"To: cliente@example.com\r\n",
		"MIME-Version: 1.0\r\n",
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n",
	} {
		if !strings.Contains(mensagem, esperado) {
			t.Fatalf("cabeçalho ausente: %q", esperado)
		}
	}

	// O assunto tem acento, e cabeçalho SMTP é ASCII: precisa ir codificado, senão
	// o cliente de e-mail mostra caracteres quebrados.
	if !strings.Contains(mensagem, "Subject: =?UTF-8?B?") {
		t.Fatalf("assunto com acento deveria ir codificado: %q", mensagem)
	}
	if strings.Contains(mensagem, "Subject: Seu veículo") {
		t.Fatal("o assunto foi para o cabeçalho sem codificação")
	}

	// Corpo separado dos cabeçalhos por linha em branco, como o RFC exige.
	cabecalho, corpo, encontrou := strings.Cut(mensagem, "\r\n\r\n")
	if !encontrou {
		t.Fatal("faltou a linha em branco entre cabeçalhos e corpo")
	}
	if strings.Contains(cabecalho, "Olá") {
		t.Fatal("o conteúdo vazou para os cabeçalhos")
	}
	if !strings.Contains(corpo, "O serviço foi concluído.") {
		t.Fatalf("corpo = %q", corpo)
	}
}

// Servidor sem autenticação — o caso do Mailpit local.
func TestNewSMTPEnviadorSemAutenticacao(t *testing.T) {
	enviador := NewSMTPEnviador("mailpit", 1025, "oficina@example.com", "", "")

	if enviador.endereco != "mailpit:1025" {
		t.Fatalf("endereco = %q", enviador.endereco)
	}
	if enviador.autenticar != nil {
		t.Fatal("sem usuário, não deveria montar autenticação")
	}
}

func TestNewSMTPEnviadorComAutenticacao(t *testing.T) {
	enviador := NewSMTPEnviador("smtp.exemplo.com", 587, "oficina@example.com", "usuario", "senha")

	if enviador.autenticar == nil {
		t.Fatal("com usuário, a autenticação deveria estar montada")
	}
}

// O envelope SMTP aceita só o endereço; mandar "Nome <a@b.com>" ali faz o servidor
// responder 501. O nome de exibição vale apenas no cabeçalho From.
func TestRemetenteDoEnvelopeNaoLevaNomeDeExibicao(t *testing.T) {
	casos := []struct {
		remetente string
		esperado  string
	}{
		{"Oficina Mecanica <oficina@example.com>", "oficina@example.com"},
		{"oficina@example.com", "oficina@example.com"},
		{"  Oficina <oficina@example.com>  ", "oficina@example.com"},
	}

	for _, caso := range casos {
		t.Run(caso.remetente, func(t *testing.T) {
			enviador := NewSMTPEnviador("mailpit", 1025, caso.remetente, "", "")
			if enviador.enderecoRemetente != caso.esperado {
				t.Fatalf("envelope = %q, esperado %q", enviador.enderecoRemetente, caso.esperado)
			}
			// O cabeçalho continua com a forma completa.
			if enviador.remetente != caso.remetente {
				t.Fatalf("cabeçalho = %q, deveria preservar %q", enviador.remetente, caso.remetente)
			}
		})
	}
}
