package orcamento

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type enviarRepositoryFake struct {
	dados      OrcamentoParaEnvio
	erroBusca  error
	erroMarcar error
	marcou     bool
	// statusEsperado guarda o que o caso de uso mandou o repositorio conferir no UPDATE.
	statusEsperado string
}

func (fake *enviarRepositoryFake) BuscarParaEnvio(context.Context, string) (OrcamentoParaEnvio, error) {
	return fake.dados, fake.erroBusca
}

func (fake *enviarRepositoryFake) MarcarEnviado(_ context.Context, _, _, statusEsperado, _ string) (time.Time, error) {
	fake.marcou = true
	fake.statusEsperado = statusEsperado
	return time.Now(), fake.erroMarcar
}

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

func mudo() *log.Logger { return log.New(os.NewFile(0, os.DevNull), "", 0) }

func dadosValidos() OrcamentoParaEnvio {
	return OrcamentoParaEnvio{
		Orcamento: orcamento.Orcamento{
			ID: principalID, Tipo: orcamento.TipoPrincipal, Status: orcamento.StatusCriado,
			Itens: []orcamento.Item{{ID: "i1", Quantidade: 1, ValorUnitario: 150}},
		},
		OrdemServicoID: ordemID,
		ClienteID:      "20000000-0000-0000-0000-000000000001",
		StatusOS:       orcamento.OSStatusEmDiagnostico,
		Calculado:      true,
	}
}

func TestEnviarMoveAOSParaAguardandoAprovacao(t *testing.T) {
	repo := &enviarRepositoryFake{dados: dadosValidos()}
	notificador := &notificadorFake{}

	resultado, err := NewEnviar(repo, notificador, mudo()).Execute(context.Background(), principalID, "usuario-1")
	if err != nil {
		t.Fatal(err)
	}
	if !repo.marcou {
		t.Fatal("o envio deveria ter marcado a OS")
	}
	if resultado.StatusOrdemServico != orcamento.OSStatusAguardandoAprovacao {
		t.Fatalf("status = %q", resultado.StatusOrdemServico)
	}
	if !resultado.NotificacaoEnviada {
		t.Fatal("a notificacao deveria ter sido enfileirada")
	}
}

func TestEnviarNotificaOClienteComOEventoCerto(t *testing.T) {
	repo := &enviarRepositoryFake{dados: dadosValidos()}
	notificador := &notificadorFake{}

	if _, err := NewEnviar(repo, notificador, mudo()).Execute(context.Background(), principalID, "usuario-1"); err != nil {
		t.Fatal(err)
	}
	if !notificador.chamado {
		t.Fatal("o cliente deveria ser avisado de que o orcamento esta pronto")
	}
	if notificador.recebido.TipoEvento != notificacaoDominio.EventoOrcamentoPronto {
		t.Fatalf("evento = %q", notificador.recebido.TipoEvento)
	}
	if notificador.recebido.Origem.Agregado != "orcamento" {
		t.Fatalf("origem = %+v", notificador.recebido.Origem)
	}
}

// Mesma regra do finalizar: falha no aviso nao desfaz a transicao da OS.
func TestEnviarSobreviveAFalhaDaNotificacao(t *testing.T) {
	repo := &enviarRepositoryFake{dados: dadosValidos()}
	notificador := &notificadorFake{erro: errors.New("cliente sem e-mail")}

	resultado, err := NewEnviar(repo, notificador, mudo()).Execute(context.Background(), principalID, "usuario-1")
	if err != nil {
		t.Fatalf("a falha da notificacao nao podia virar erro do envio: %v", err)
	}
	if !repo.marcou {
		t.Fatal("a OS precisa ter transitado mesmo assim")
	}
	if resultado.NotificacaoEnviada {
		t.Fatal("NotificacaoEnviada deveria ser false")
	}
}

func TestEnviarRejeitaIdentificadorInvalido(t *testing.T) {
	repo := &enviarRepositoryFake{}
	if _, err := NewEnviar(repo, &notificadorFake{}, mudo()).Execute(context.Background(), "nao-e-uuid", ""); !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v", err)
	}
	if repo.marcou {
		t.Fatal("nada deveria ser marcado")
	}
}

func TestEnviarPropagaRegrasDeDominio(t *testing.T) {
	casos := []struct {
		nome     string
		ajustar  func(*OrcamentoParaEnvio)
		esperado error
	}{
		{"nao calculado", func(d *OrcamentoParaEnvio) { d.Calculado = false }, orcamento.ErrOrcamentoNaoCalculado},
		{"ja enviado", func(d *OrcamentoParaEnvio) { d.StatusOS = orcamento.OSStatusAguardandoAprovacao }, orcamento.ErrOrcamentoJaEnviado},
		{"sem itens", func(d *OrcamentoParaEnvio) { d.Orcamento.Itens = nil }, orcamento.ErrSemItens},
		{"OS em estado invalido", func(d *OrcamentoParaEnvio) { d.StatusOS = "RECEBIDA" }, orcamento.ErrOSNaoPermiteEnvio},
		{"orcamento aprovado", func(d *OrcamentoParaEnvio) { d.Orcamento.Status = orcamento.StatusAprovado }, orcamento.ErrStatusNaoCalculavel},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			dados := dadosValidos()
			caso.ajustar(&dados)
			repo := &enviarRepositoryFake{dados: dados}

			_, err := NewEnviar(repo, &notificadorFake{}, mudo()).Execute(context.Background(), principalID, "u")
			if !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
			if repo.marcou {
				t.Fatal("a OS nao podia ter transitado")
			}
		})
	}
}

// A gravação precisa exigir o mesmo status que a validação observou: entre uma e outra a
// OS pode ter mudado, e dois envios simultâneos mandariam dois e-mails ao cliente.
func TestEnviarExigeNoUpdateOStatusQueValidou(t *testing.T) {
	repo := &enviarRepositoryFake{dados: dadosValidos()}

	if _, err := NewEnviar(repo, &notificadorFake{}, mudo()).Execute(context.Background(), principalID, "usuario-1"); err != nil {
		t.Fatal(err)
	}
	if repo.statusEsperado != orcamento.OSStatusEmDiagnostico {
		t.Fatalf("statusEsperado = %q; o UPDATE precisa condicionar ao status validado", repo.statusEsperado)
	}
}

// Quem perde a corrida não avisa o cliente: a OS não foi marcada por este envio.
func TestEnviarQuePerdeACorridaNaoNotifica(t *testing.T) {
	notificador := &notificadorFake{}
	repo := &enviarRepositoryFake{dados: dadosValidos(), erroMarcar: orcamento.ErrOSNaoPermiteEnvio}

	_, err := NewEnviar(repo, notificador, mudo()).Execute(context.Background(), principalID, "usuario-1")
	if !errors.Is(err, orcamento.ErrOSNaoPermiteEnvio) {
		t.Fatalf("err = %v; o conflito precisa chegar ao cliente da API", err)
	}
	if notificador.chamado {
		t.Fatal("não se envia e-mail de um orçamento que outro envio já marcou")
	}
}
