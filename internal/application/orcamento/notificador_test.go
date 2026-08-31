package orcamento

import (
	"context"
	"errors"
	"testing"

	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

const (
	osIDNotificacao      = "70000000-0000-0000-0000-000000000001"
	clienteIDNotificacao = "20000000-0000-0000-0000-000000000001"
	orcamentoIDValido    = "b86b2ba5-b896-4557-85f2-55e1a5ba8ba3"
)

func aprovacao(statusOS string) domain.Aprovacao {
	return domain.Aprovacao{
		OrcamentoID: orcamentoIDValido, OrdemServicoID: osIDNotificacao,
		ClienteID: clienteIDNotificacao, StatusOrdemServico: statusOS,
	}
}

// Aprovar leva a dois destinos, e o cliente precisa saber a qual deles chegou.
func TestAprovarEscolheOEventoPeloDestinoDaOS(t *testing.T) {
	casos := []struct {
		statusOS string
		evento   string
	}{
		{"AGUARDANDO_EXECUCAO", notificacaoDominio.EventoOrcamentoAprovado},
		{"AGUARDANDO_RECURSOS", notificacaoDominio.EventoAguardandoRecursos},
	}
	for _, caso := range casos {
		t.Run(caso.statusOS, func(t *testing.T) {
			notificador := &notificadorFake{}
			repo := &aprovarRepositoryFake{resultado: aprovacao(caso.statusOS)}

			_, err := NewAprovar(repo, notificador, mudo()).Execute(context.Background(), AprovarInput{
				OrcamentoID: orcamentoIDValido, ClienteID: clienteIDNotificacao,
			})
			if err != nil {
				t.Fatal(err)
			}
			if notificador.recebido.TipoEvento != caso.evento {
				t.Fatalf("evento = %q; esperado %q", notificador.recebido.TipoEvento, caso.evento)
			}
			if notificador.recebido.ClienteID != clienteIDNotificacao {
				t.Fatalf("clienteID = %q", notificador.recebido.ClienteID)
			}
			if notificador.recebido.Origem.ID != osIDNotificacao {
				t.Fatalf("origem = %+v", notificador.recebido.Origem)
			}
		})
	}
}

// RNF-OS-44: a aprovação já reservou estoque e não pode ser desfeita pelo e-mail.
func TestAprovarSobreviveAFalhaDaNotificacao(t *testing.T) {
	notificador := &notificadorFake{erro: errors.New("cliente sem e-mail")}
	repo := &aprovarRepositoryFake{resultado: aprovacao("AGUARDANDO_EXECUCAO")}

	resultado, err := NewAprovar(repo, notificador, mudo()).Execute(context.Background(), AprovarInput{
		OrcamentoID: orcamentoIDValido, ClienteID: clienteIDNotificacao,
	})
	if err != nil {
		t.Fatalf("a falha na notificação não podia virar erro da aprovação: %v", err)
	}
	if resultado.StatusOrdemServico != "AGUARDANDO_EXECUCAO" {
		t.Fatalf("status = %q", resultado.StatusOrdemServico)
	}
}

func TestAprovarComErroNaoNotifica(t *testing.T) {
	notificador := &notificadorFake{}
	repo := &aprovarRepositoryFake{err: ErrOrcamentoJaDecidido}

	if _, err := NewAprovar(repo, notificador, mudo()).Execute(context.Background(), AprovarInput{
		OrcamentoID: orcamentoIDValido, ClienteID: clienteIDNotificacao,
	}); err == nil {
		t.Fatal("o erro do repositório deveria ser propagado")
	}
	if notificador.chamado {
		t.Fatal("não se avisa aprovação que não aconteceu")
	}
}

// OS -> CANCELADA
func TestRecusarPrincipalAvisaOCancelamento(t *testing.T) {
	notificador := &notificadorFake{}
	stub := &recusarRepositoryStub{result: domain.Decisao{
		OrcamentoID: orcamentoIDValido, OrdemServicoID: osIDNotificacao,
		ClienteID: clienteIDNotificacao, StatusOrdemServico: "CANCELADA",
	}}

	if _, err := NewRecusar(stub, notificador, mudo()).Execute(context.Background(), RecusarInput{
		OrcamentoID: orcamentoIDValido,
	}); err != nil {
		t.Fatal(err)
	}
	if notificador.recebido.TipoEvento != notificacaoDominio.EventoServicoCancelado {
		t.Fatalf("evento = %q", notificador.recebido.TipoEvento)
	}
}

// A recusa de um COMPLEMENTAR devolve a OS para AGUARDANDO_EXECUCAO. É consequência
// direta do que o cliente acabou de decidir: não há novidade a comunicar.
func TestRecusarComplementarNaoAvisa(t *testing.T) {
	notificador := &notificadorFake{}
	stub := &recusarRepositoryStub{result: domain.Decisao{
		OrcamentoID: orcamentoIDValido, OrdemServicoID: osIDNotificacao,
		ClienteID: clienteIDNotificacao, StatusOrdemServico: "AGUARDANDO_EXECUCAO",
	}}

	if _, err := NewRecusar(stub, notificador, mudo()).Execute(context.Background(), RecusarInput{
		OrcamentoID: orcamentoIDValido,
	}); err != nil {
		t.Fatal(err)
	}
	if notificador.chamado {
		t.Fatal("a OS voltou para a fila por decisão do próprio cliente; não há o que avisar")
	}
}

func TestRecusarSobreviveAFalhaDaNotificacao(t *testing.T) {
	notificador := &notificadorFake{erro: errors.New("smtp fora do ar")}
	stub := &recusarRepositoryStub{result: domain.Decisao{
		OrcamentoID: orcamentoIDValido, OrdemServicoID: osIDNotificacao,
		ClienteID: clienteIDNotificacao, StatusOrdemServico: "CANCELADA",
	}}

	resultado, err := NewRecusar(stub, notificador, mudo()).Execute(context.Background(), RecusarInput{
		OrcamentoID: orcamentoIDValido,
	})
	if err != nil {
		t.Fatalf("a falha na notificação não podia virar erro da recusa: %v", err)
	}
	if resultado.StatusOrdemServico != "CANCELADA" {
		t.Fatalf("status = %q; a OS precisa continuar CANCELADA", resultado.StatusOrdemServico)
	}
}
