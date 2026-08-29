package ordemservico

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

const (
	osID      = "70000000-0000-0000-0000-000000000001"
	clienteID = "20000000-0000-0000-0000-000000000001"
)

type notificadorFake struct {
	recebido notificacaoApplication.Pedido
	chamado  bool
	erro     error
}

func (fake *notificadorFake) Execute(_ context.Context, pedido notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error) {
	fake.chamado = true
	fake.recebido = pedido
	return notificacaoDominio.Notificacao{}, fake.erro
}

type finalizarFake struct {
	resultado domain.ResultadoFinalizacao
	erro      error
}

func (fake finalizarFake) Finalizar(context.Context, FinalizarInput) (domain.ResultadoFinalizacao, error) {
	return fake.resultado, fake.erro
}

func silencioso() *log.Logger { return log.New(os.NewFile(0, os.DevNull), "", 0) }

func TestFinalizarEnfileiraNotificacao(t *testing.T) {
	notificador := &notificadorFake{}
	repo := finalizarFake{resultado: domain.ResultadoFinalizacao{
		OrdemServicoID: osID, ClienteID: clienteID, Status: domain.StatusFinalizada,
	}}

	resultado, err := NewFinalizar(repo, notificador, silencioso()).Execute(context.Background(), FinalizarInput{OSID: osID})
	if err != nil {
		t.Fatal(err)
	}
	if !notificador.chamado {
		t.Fatal("a finalização deveria enfileirar a notificação (RF-OS-87)")
	}
	if notificador.recebido.TipoEvento != notificacaoDominio.EventoServicoFinalizado {
		t.Fatalf("evento = %q", notificador.recebido.TipoEvento)
	}
	if notificador.recebido.ClienteID != clienteID {
		t.Fatalf("clienteID = %q", notificador.recebido.ClienteID)
	}
	if notificador.recebido.Origem.Agregado != "ordem_servico" || notificador.recebido.Origem.ID != osID {
		t.Fatalf("origem = %+v", notificador.recebido.Origem)
	}
	if !resultado.NotificacaoEnviada {
		t.Fatal("NotificacaoEnviada deveria refletir o enfileiramento (RF-OS-88)")
	}
}

// RNF-OS-44: falha ao notificar não pode desfazer a finalização.
func TestFinalizarSobreviveAFalhaDaNotificacao(t *testing.T) {
	notificador := &notificadorFake{erro: errors.New("cliente sem e-mail")}
	repo := finalizarFake{resultado: domain.ResultadoFinalizacao{
		OrdemServicoID: osID, ClienteID: clienteID, Status: domain.StatusFinalizada,
	}}

	resultado, err := NewFinalizar(repo, notificador, silencioso()).Execute(context.Background(), FinalizarInput{OSID: osID})
	if err != nil {
		t.Fatalf("a falha na notificação não podia virar erro da finalização: %v", err)
	}
	if resultado.Status != domain.StatusFinalizada {
		t.Fatalf("status = %q; a OS precisa continuar FINALIZADA", resultado.Status)
	}
	if resultado.NotificacaoEnviada {
		t.Fatal("NotificacaoEnviada deveria ser false quando o enfileiramento falha")
	}
}

// Quando a finalização falha, nada é notificado — não há o que avisar.
func TestFinalizarComErroNaoNotifica(t *testing.T) {
	notificador := &notificadorFake{}
	repo := finalizarFake{erro: domain.ErrOSNaoEmExecucao}

	if _, err := NewFinalizar(repo, notificador, silencioso()).Execute(context.Background(), FinalizarInput{OSID: osID}); err == nil {
		t.Fatal("o erro da finalização deveria ser propagado")
	}
	if notificador.chamado {
		t.Fatal("não se notifica cliente de uma finalização que não aconteceu")
	}
}

// Sem notificador configurado a finalização segue normalmente.
func TestFinalizarSemNotificador(t *testing.T) {
	repo := finalizarFake{resultado: domain.ResultadoFinalizacao{OrdemServicoID: osID, ClienteID: clienteID}}

	resultado, err := NewFinalizar(repo, nil, silencioso()).Execute(context.Background(), FinalizarInput{OSID: osID})
	if err != nil {
		t.Fatal(err)
	}
	if resultado.NotificacaoEnviada {
		t.Fatal("sem notificador, NotificacaoEnviada precisa ser false")
	}
}

type entregarFake struct {
	resultado domain.ResultadoEntrega
	erro      error
}

func (fake entregarFake) Entregar(context.Context, EntregarInput) (domain.ResultadoEntrega, error) {
	return fake.resultado, fake.erro
}

func TestEntregarEnfileiraNotificacao(t *testing.T) {
	notificador := &notificadorFake{}
	repo := entregarFake{resultado: domain.ResultadoEntrega{OrdemServicoID: osID, ClienteID: clienteID}}

	if _, err := NewEntregar(repo, notificador, silencioso()).Execute(context.Background(), EntregarInput{OSID: osID}); err != nil {
		t.Fatal(err)
	}
	if !notificador.chamado {
		t.Fatal("a entrega deveria enfileirar a confirmação")
	}
	if notificador.recebido.TipoEvento != notificacaoDominio.EventoVeiculoEntregue {
		t.Fatalf("evento = %q", notificador.recebido.TipoEvento)
	}
}

func TestEntregarSobreviveAFalhaDaNotificacao(t *testing.T) {
	notificador := &notificadorFake{erro: errors.New("smtp fora do ar")}
	repo := entregarFake{resultado: domain.ResultadoEntrega{OrdemServicoID: osID, ClienteID: clienteID}}

	if _, err := NewEntregar(repo, notificador, silencioso()).Execute(context.Background(), EntregarInput{OSID: osID}); err != nil {
		t.Fatalf("a falha na notificação não podia virar erro da entrega: %v", err)
	}
}
