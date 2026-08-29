package notificacao

import (
	"context"
	"errors"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

var ErrClienteNaoEncontrado = errors.New("cliente nao encontrado")

// Contato e o que o modulo notificador precisa saber do cliente. Fica aqui, e nao no
// pacote cliente, para nao acoplar os contextos.
type Contato struct {
	ID    string
	Nome  string
	Email string
}

type Repository interface {
	// ContatoDoCliente devolve os dados de contato usados no envio.
	ContatoDoCliente(ctx context.Context, clienteID string) (Contato, error)
	// Enfileirar grava a notificacao como PENDENTE.
	Enfileirar(ctx context.Context, nova notificacao.Notificacao) (notificacao.Notificacao, error)
}

// Enfileirar e a porta que os demais contextos usam. Ela nao envia nada: registra a
// intencao e retorna. O envio e responsabilidade do processador da fila, e e isso que
// garante que uma falha de e-mail nunca desfaca a operacao de negocio (RNF-OS-44).
type Enfileirar struct {
	repository Repository
}

func NewEnfileirar(repository Repository) Enfileirar {
	return Enfileirar{repository: repository}
}

// Pedido descreve o aviso desejado sem exigir que o chamador conheca o e-mail do cliente
// nem o formato da mensagem.
type Pedido struct {
	ClienteID  string
	TipoEvento string
	Origem     notificacao.Origem
}

func (useCase Enfileirar) Execute(ctx context.Context, pedido Pedido) (notificacao.Notificacao, error) {
	contato, err := useCase.repository.ContatoDoCliente(ctx, pedido.ClienteID)
	if err != nil {
		return notificacao.Notificacao{}, err
	}

	assunto, conteudo, err := Mensagem(pedido.TipoEvento, contato.Nome)
	if err != nil {
		return notificacao.Notificacao{}, err
	}

	nova, err := notificacao.Nova(pedido.ClienteID, contato.Email, pedido.TipoEvento, assunto, conteudo, pedido.Origem)
	if err != nil {
		return notificacao.Notificacao{}, err
	}
	return useCase.repository.Enfileirar(ctx, nova)
}
