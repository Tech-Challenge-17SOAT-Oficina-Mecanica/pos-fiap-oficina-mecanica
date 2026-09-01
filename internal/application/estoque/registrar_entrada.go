package estoque

import (
	"context"
	"errors"
	"log"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrItemNaoEncontrado         = errors.New("item de estoque nao encontrado")
	ErrItemInativo               = errors.New("item de estoque inativo")
	ErrFornecedorNaoEncontrado   = errors.New("fornecedor nao encontrado")
	ErrFornecedorInativo         = errors.New("fornecedor inativo")
	ErrFornecedorDivergente      = errors.New("fornecedor do pedido de compra diverge do recebimento")
	ErrDocumentoOrigemDuplicado  = errors.New("documentoOrigem ja registrado")
	ErrPedidoCompraNaoEncontrado = errors.New("pedido de compra nao encontrado")
	ErrItemForaDoPedido          = errors.New("item nao pertence ao pedido de compra informado")
	ErrDivergenciaQuantidade     = errors.New("quantidade recebida maior que a pedida; confirme a divergencia")
)

type ItemInput struct {
	ItemID        string
	Quantidade    float64
	CustoUnitario float64
}

type RegistrarEntradaInput struct {
	IdempotencyKey       string
	DocumentoOrigem      string
	FornecedorID         string
	PedidoCompraID       string
	ConfirmarDivergencia bool
	Itens                []ItemInput
	UsuarioID            string
}

// Resultado agrupa o retorno do repositorio com o indicador de repeticao por idempotencia.
type Resultado struct {
	Entrada      domain.ResultadoEntrada
	JaProcessada bool
}

type EntradaRepository interface {
	RegistrarEntrada(context.Context, RegistrarEntradaInput, domain.EntradaCadastro) (Resultado, error)
}

type RegistrarEntrada struct {
	repository  EntradaRepository
	notificador Notificador
	logger      *log.Logger
}

func NewRegistrarEntrada(repository EntradaRepository, notificador Notificador, logger *log.Logger) RegistrarEntrada {
	if logger == nil {
		logger = log.Default()
	}
	return RegistrarEntrada{repository: repository, notificador: notificador, logger: logger}
}

func (useCase RegistrarEntrada) Execute(ctx context.Context, input RegistrarEntradaInput) (Resultado, error) {
	if !validation.IsUUID(input.IdempotencyKey) {
		return Resultado{}, domain.ErrIdempotencyKeyObrigatoria
	}
	if input.FornecedorID != "" && !validation.IsUUID(input.FornecedorID) {
		return Resultado{}, domain.ErrFornecedorIDInvalido
	}
	itens := make([]domain.ItemEntrada, 0, len(input.Itens))
	for _, item := range input.Itens {
		if !validation.IsUUID(item.ItemID) {
			return Resultado{}, domain.ErrItemIDInvalido
		}
		itens = append(itens, domain.ItemEntrada{ItemID: item.ItemID, Quantidade: item.Quantidade, CustoUnitario: item.CustoUnitario})
	}
	cadastro, err := domain.NovaEntradaCadastro(input.DocumentoOrigem, input.FornecedorID, input.PedidoCompraID, input.ConfirmarDivergencia, itens)
	if err != nil {
		return Resultado{}, err
	}
	resultado, err := useCase.repository.RegistrarEntrada(ctx, input, cadastro)
	if err != nil {
		return Resultado{}, err
	}

	// A repeticao por idempotencia devolve a resposta guardada sem mexer em nada: avisar
	// de novo mandaria o mesmo e-mail ao cliente por causa de um retry da integracao.
	if !resultado.JaProcessada {
		useCase.avisarOrdensLiberadas(ctx, resultado.Entrada.OrdensServico)
	}
	return resultado, nil
}

// avisarOrdensLiberadas comunica o cliente cuja OS saiu da espera por pecas. Acontece fora
// da transacao, e uma OS que continua em AGUARDANDO_RECURSOS nao gera aviso: para o
// cliente nada mudou, porque ainda faltam itens.
func (useCase RegistrarEntrada) avisarOrdensLiberadas(ctx context.Context, ordens []domain.OrdemServicoLiberada) {
	for _, ordem := range ordens {
		if ordem.Status != "AGUARDANDO_EXECUCAO" || ordem.Status == ordem.StatusAnterior {
			continue
		}
		avisar(ctx, useCase.notificador, ordem.ClienteID,
			notificacaoDominio.EventoRecursosDisponiveis, ordem.OrdemServicoID,
			func(erro error) {
				useCase.logger.Printf("notificacao de liberacao da OS %s nao pode ser enfileirada: %v", ordem.OrdemServicoID, erro)
			})
	}
}
