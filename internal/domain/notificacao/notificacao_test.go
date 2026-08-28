package notificacao

import (
	"errors"
	"testing"
	"time"
)

const (
	clienteID = "20000000-0000-0000-0000-000000000001"
	osID      = "70000000-0000-0000-0000-000000000001"
)

func origemValida() Origem { return Origem{Agregado: "ordem_servico", ID: osID} }

func TestNovaNasceCendente(t *testing.T) {
	aviso, err := Nova(clienteID, "  cliente@example.com  ", EventoServicoFinalizado,
		" Veículo pronto ", " Seu carro está pronto. ", origemValida())
	if err != nil {
		t.Fatal(err)
	}
	if aviso.Status != StatusPendente {
		t.Fatalf("status = %q, esperado PENDENTE", aviso.Status)
	}
	if aviso.Canal != CanalEmail {
		t.Fatalf("canal = %q", aviso.Canal)
	}
	if aviso.Destinatario != "cliente@example.com" {
		t.Fatalf("destinatario = %q, deveria estar sem espaços", aviso.Destinatario)
	}
	if aviso.Tentativas != 0 || aviso.EnviadaEm != nil {
		t.Fatal("uma notificação nova não pode ter tentativa nem data de envio")
	}
}

func TestNovaValidacoes(t *testing.T) {
	casos := []struct {
		nome     string
		executar func() (Notificacao, error)
		esperado error
	}{
		{"cliente não uuid", func() (Notificacao, error) {
			return Nova("abc", "a@b.com", EventoServicoFinalizado, "a", "b", origemValida())
		}, ErrClienteObrigatorio},
		{"sem e-mail", func() (Notificacao, error) {
			return Nova(clienteID, "  ", EventoServicoFinalizado, "a", "b", origemValida())
		}, ErrDestinatarioObrigatorio},
		{"e-mail inválido", func() (Notificacao, error) {
			return Nova(clienteID, "nao-e-email", EventoServicoFinalizado, "a", "b", origemValida())
		}, ErrDestinatarioInvalido},
		{"evento desconhecido", func() (Notificacao, error) {
			return Nova(clienteID, "a@b.com", "QUALQUER_COISA", "a", "b", origemValida())
		}, ErrEventoInvalido},
		{"origem sem agregado", func() (Notificacao, error) {
			return Nova(clienteID, "a@b.com", EventoServicoFinalizado, "a", "b", Origem{ID: osID})
		}, ErrAgregadoObrigatorio},
		{"origem com id inválido", func() (Notificacao, error) {
			return Nova(clienteID, "a@b.com", EventoServicoFinalizado, "a", "b", Origem{Agregado: "os", ID: "abc"})
		}, ErrAgregadoObrigatorio},
		{"assunto vazio", func() (Notificacao, error) {
			return Nova(clienteID, "a@b.com", EventoServicoFinalizado, "  ", "b", origemValida())
		}, ErrAssuntoObrigatorio},
		{"conteúdo vazio", func() (Notificacao, error) {
			return Nova(clienteID, "a@b.com", EventoServicoFinalizado, "a", "  ", origemValida())
		}, ErrConteudoObrigatorio},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if _, err := caso.executar(); !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
		})
	}
}

func TestMarcarEnviada(t *testing.T) {
	aviso, _ := Nova(clienteID, "a@b.com", EventoServicoFinalizado, "a", "b", origemValida())
	momento := time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC)

	enviada, err := aviso.MarcarEnviada(momento)
	if err != nil {
		t.Fatal(err)
	}
	if enviada.Status != StatusEnviada || enviada.Tentativas != 1 {
		t.Fatalf("status = %q, tentativas = %d", enviada.Status, enviada.Tentativas)
	}
	if enviada.EnviadaEm == nil || !enviada.EnviadaEm.Equal(momento) {
		t.Fatal("a data de envio deveria ter sido registrada")
	}
}

// Reenviar algo já entregue duplicaria o aviso ao cliente.
func TestMarcarEnviadaDuasVezesEhRecusado(t *testing.T) {
	aviso, _ := Nova(clienteID, "a@b.com", EventoServicoFinalizado, "a", "b", origemValida())
	enviada, _ := aviso.MarcarEnviada(time.Now())

	if _, err := enviada.MarcarEnviada(time.Now()); !errors.Is(err, ErrJaEnviada) {
		t.Fatalf("erro = %v, esperado %v", err, ErrJaEnviada)
	}
}

func TestMarcarFalhaMantemReenviavel(t *testing.T) {
	aviso, _ := Nova(clienteID, "a@b.com", EventoServicoFinalizado, "a", "b", origemValida())

	falhou := aviso.MarcarFalha("smtp indisponível")
	if falhou.Status != StatusFalhou || falhou.Tentativas != 1 {
		t.Fatalf("status = %q, tentativas = %d", falhou.Status, falhou.Tentativas)
	}
	if falhou.UltimoErro == nil || *falhou.UltimoErro != "smtp indisponível" {
		t.Fatal("o motivo da falha deveria ter sido guardado")
	}
	if !falhou.Reenviavel() {
		t.Fatal("uma notificação que falhou precisa continuar reenviável")
	}
	if falhou.EnviadaEm != nil {
		t.Fatal("falha não pode registrar data de envio")
	}
}

func TestReenviavel(t *testing.T) {
	casos := map[string]bool{StatusPendente: true, StatusFalhou: true, StatusEnviada: false}
	for status, esperado := range casos {
		if obtido := (Notificacao{Status: status}).Reenviavel(); obtido != esperado {
			t.Fatalf("status %q reenviável = %v, esperado %v", status, obtido, esperado)
		}
	}
}
