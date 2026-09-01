package estoque

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

const (
	osLiberadaID = "70000000-0000-0000-0000-000000000001"
	osRetidaID   = "70000000-0000-0000-0000-000000000002"
	clienteID    = "20000000-0000-0000-0000-000000000001"
)

type notificadorFake struct {
	recebidos []notificacaoApplication.Pedido
	erro      error
}

func (fake *notificadorFake) Execute(_ context.Context, pedido notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error) {
	fake.recebidos = append(fake.recebidos, pedido)
	return notificacaoDominio.Notificacao{}, fake.erro
}

func mudo() *log.Logger { return log.New(os.NewFile(0, os.DevNull), "", 0) }

func entradaValida() RegistrarEntradaInput {
	return RegistrarEntradaInput{
		IdempotencyKey:  "11111111-1111-1111-1111-111111111111",
		DocumentoOrigem: "NF-1234",
		Itens: []ItemInput{{
			ItemID: "30000000-0000-0000-0000-000000000001", Quantidade: 2, CustoUnitario: 10,
		}},
	}
}

func resultadoCom(ordens []domain.OrdemServicoLiberada, jaProcessada bool) Resultado {
	return Resultado{
		Entrada:      domain.ResultadoEntrada{EntradaID: "entrada", OrdensServico: ordens},
		JaProcessada: jaProcessada,
	}
}

// A chegada das peças tira a OS da espera: é a única situação em que o cliente tem
// novidade. Uma OS que continua faltando item não gera aviso.
func TestRegistrarEntradaAvisaApenasAsOrdensLiberadas(t *testing.T) {
	notificador := &notificadorFake{}
	repo := &registrarEntradaRepositoryFake{resultado: resultadoCom([]domain.OrdemServicoLiberada{
		{OrdemServicoID: osLiberadaID, ClienteID: clienteID, StatusAnterior: "AGUARDANDO_RECURSOS", Status: "AGUARDANDO_EXECUCAO"},
		{OrdemServicoID: osRetidaID, ClienteID: clienteID, StatusAnterior: "AGUARDANDO_RECURSOS", Status: "AGUARDANDO_RECURSOS", ItensPendentes: 1},
	}, false)}

	if _, err := NewRegistrarEntrada(repo, notificador, mudo()).Execute(context.Background(), entradaValida()); err != nil {
		t.Fatal(err)
	}
	if len(notificador.recebidos) != 1 {
		t.Fatalf("avisos = %d; só a OS liberada devia ser notificada", len(notificador.recebidos))
	}
	aviso := notificador.recebidos[0]
	if aviso.TipoEvento != notificacaoDominio.EventoRecursosDisponiveis {
		t.Fatalf("evento = %q", aviso.TipoEvento)
	}
	if aviso.Origem.ID != osLiberadaID || aviso.ClienteID != clienteID {
		t.Fatalf("pedido = %+v", aviso)
	}
}

// O retry de uma integração devolve a resposta guardada; reavisar mandaria o mesmo
// e-mail de novo ao cliente.
func TestRegistrarEntradaIdempotenteNaoReavisa(t *testing.T) {
	notificador := &notificadorFake{}
	repo := &registrarEntradaRepositoryFake{resultado: resultadoCom([]domain.OrdemServicoLiberada{
		{OrdemServicoID: osLiberadaID, ClienteID: clienteID, StatusAnterior: "AGUARDANDO_RECURSOS", Status: "AGUARDANDO_EXECUCAO"},
	}, true)}

	if _, err := NewRegistrarEntrada(repo, notificador, mudo()).Execute(context.Background(), entradaValida()); err != nil {
		t.Fatal(err)
	}
	if len(notificador.recebidos) != 0 {
		t.Fatalf("avisos = %d; a repetição por idempotência não notifica", len(notificador.recebidos))
	}
}

// RNF-OS-44: a entrada de estoque já está gravada e não pode ser desfeita pelo e-mail.
func TestRegistrarEntradaSobreviveAFalhaDaNotificacao(t *testing.T) {
	notificador := &notificadorFake{erro: errors.New("cliente sem e-mail")}
	repo := &registrarEntradaRepositoryFake{resultado: resultadoCom([]domain.OrdemServicoLiberada{
		{OrdemServicoID: osLiberadaID, ClienteID: clienteID, StatusAnterior: "AGUARDANDO_RECURSOS", Status: "AGUARDANDO_EXECUCAO"},
	}, false)}

	resultado, err := NewRegistrarEntrada(repo, notificador, mudo()).Execute(context.Background(), entradaValida())
	if err != nil {
		t.Fatalf("a falha na notificação não podia virar erro da entrada: %v", err)
	}
	if resultado.Entrada.EntradaID != "entrada" {
		t.Fatalf("a entrada precisa continuar registrada: %+v", resultado)
	}
}
