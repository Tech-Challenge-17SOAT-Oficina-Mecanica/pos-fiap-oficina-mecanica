package estoque

import (
	"context"
	"errors"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

var (
	ErrItemNaoEncontrado         = errors.New("item de estoque nao encontrado")
	ErrItemInativo               = errors.New("item de estoque inativo")
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

type RegistrarEntrada struct{ repository EntradaRepository }

func NewRegistrarEntrada(repository EntradaRepository) RegistrarEntrada {
	return RegistrarEntrada{repository: repository}
}

func (useCase RegistrarEntrada) Execute(ctx context.Context, input RegistrarEntradaInput) (Resultado, error) {
	if !validation.IsUUID(input.IdempotencyKey) {
		return Resultado{}, domain.ErrIdempotencyKeyObrigatoria
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
	return useCase.repository.RegistrarEntrada(ctx, input, cadastro)
}
