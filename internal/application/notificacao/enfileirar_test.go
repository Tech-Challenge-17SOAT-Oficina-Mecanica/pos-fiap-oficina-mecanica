package notificacao

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

const (
	clienteID = "20000000-0000-0000-0000-000000000001"
	osID      = "70000000-0000-0000-0000-000000000001"
)

type repositoryFake struct {
	contato      Contato
	erroContato  error
	enfileirada  notificacao.Notificacao
	enfileirou   bool
	erroEnfileir error
}

func (fake *repositoryFake) ContatoDoCliente(context.Context, string) (Contato, error) {
	return fake.contato, fake.erroContato
}

func (fake *repositoryFake) Enfileirar(_ context.Context, nova notificacao.Notificacao) (notificacao.Notificacao, error) {
	fake.enfileirou = true
	fake.enfileirada = nova
	nova.ID = "10000000-0000-0000-0000-0000000000aa"
	return nova, fake.erroEnfileir
}

func pedido() Pedido {
	return Pedido{
		ClienteID:  clienteID,
		TipoEvento: notificacao.EventoServicoFinalizado,
		Origem:     notificacao.Origem{Agregado: "ordem_servico", ID: osID},
	}
}

func TestEnfileirarMontaAMensagemAPartirDoEvento(t *testing.T) {
	fake := &repositoryFake{contato: Contato{ID: clienteID, Nome: "Maria", Email: "maria@example.com"}}

	aviso, err := NewEnfileirar(fake).Execute(context.Background(), pedido())
	if err != nil {
		t.Fatal(err)
	}
	if !fake.enfileirou {
		t.Fatal("a notificação deveria ter sido enfileirada")
	}
	if aviso.Status != notificacao.StatusPendente {
		t.Fatalf("status = %q; enfileirar não envia, só registra a intenção", aviso.Status)
	}
	if fake.enfileirada.Destinatario != "maria@example.com" {
		t.Fatalf("destinatario = %q", fake.enfileirada.Destinatario)
	}
	if fake.enfileirada.Assunto == "" || fake.enfileirada.Conteudo == "" {
		t.Fatal("assunto e conteúdo deveriam ter sido montados pelo próprio módulo")
	}
}

// Quem dispara não precisa conhecer o e-mail nem o texto — só o que aconteceu.
func TestEnfileirarPersonalizaComONomeDoCliente(t *testing.T) {
	fake := &repositoryFake{contato: Contato{ID: clienteID, Nome: "Maria", Email: "maria@example.com"}}

	if _, err := NewEnfileirar(fake).Execute(context.Background(), pedido()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fake.enfileirada.Conteudo, "Maria") {
		t.Fatalf("conteúdo deveria citar o cliente: %q", fake.enfileirada.Conteudo)
	}
}

func TestEnfileirarSemEmailFalha(t *testing.T) {
	fake := &repositoryFake{contato: Contato{ID: clienteID, Nome: "Maria", Email: ""}}

	_, err := NewEnfileirar(fake).Execute(context.Background(), pedido())
	if !errors.Is(err, notificacao.ErrDestinatarioObrigatorio) {
		t.Fatalf("erro = %v, esperado %v", err, notificacao.ErrDestinatarioObrigatorio)
	}
}

func TestEnfileirarClienteInexistente(t *testing.T) {
	fake := &repositoryFake{erroContato: ErrClienteNaoEncontrado}

	if _, err := NewEnfileirar(fake).Execute(context.Background(), pedido()); !errors.Is(err, ErrClienteNaoEncontrado) {
		t.Fatalf("erro = %v", err)
	}
	if fake.enfileirou {
		t.Fatal("nada deveria ser enfileirado sem cliente")
	}
}

func TestEnfileirarEventoDesconhecido(t *testing.T) {
	fake := &repositoryFake{contato: Contato{ID: clienteID, Nome: "Maria", Email: "maria@example.com"}}
	invalido := pedido()
	invalido.TipoEvento = "EVENTO_QUE_NAO_EXISTE"

	if _, err := NewEnfileirar(fake).Execute(context.Background(), invalido); !errors.Is(err, notificacao.ErrEventoInvalido) {
		t.Fatalf("erro = %v", err)
	}
	if fake.enfileirou {
		t.Fatal("evento desconhecido não pode gerar notificação")
	}
}

func TestMensagemCobreTodosOsEventos(t *testing.T) {
	for _, evento := range []string{
		notificacao.EventoServicoFinalizado,
		notificacao.EventoOrcamentoPronto,
		notificacao.EventoVeiculoEntregue,
	} {
		t.Run(evento, func(t *testing.T) {
			assunto, conteudo, _, err := Mensagem(evento, "Maria", nil)
			if err != nil {
				t.Fatal(err)
			}
			if assunto == "" || conteudo == "" {
				t.Fatalf("evento %q ficou sem texto", evento)
			}
		})
	}
}

// Sem nome cadastrado o texto ainda precisa fazer sentido.
func TestMensagemSemNome(t *testing.T) {
	_, conteudo, _, err := Mensagem(notificacao.EventoServicoFinalizado, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(conteudo, ", !") || conteudo == "" {
		t.Fatalf("saudação ficou quebrada sem o nome: %q", conteudo)
	}
}
