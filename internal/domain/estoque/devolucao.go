package estoque

import "errors"

const (
	ReservaAtiva    = "ATIVA"
	ReservaLiberada = "LIBERADA"

	MovimentacaoLiberacaoReserva = "LIBERACAO_RESERVA"
	MovimentacaoEntradaRetorno   = "ENTRADA_RETORNO"

	MotivoPedidoDeCompraNaoRecebido = "PEDIDO_DE_COMPRA_NAO_RECEBIDO"
)

var ErrSaldoReservadoInsuficiente = errors.New("saldo reservado insuficiente para liberar a reserva")

// ItemLiberado é um item cuja reserva ativa foi liberada, sem alterar o saldo físico.
type ItemLiberado struct {
	ItemID             string
	Codigo             string
	Descricao          string
	Tipo               string
	UnidadeMedida      string
	Quantidade         float64
	SaldoReservadoApos float64
}

// ItemRetornado é um item cuja quantidade já baixada voltou ao saldo físico.
type ItemRetornado struct {
	ItemID          string
	Codigo          string
	Descricao       string
	Tipo            string
	UnidadeMedida   string
	Quantidade      float64
	SaldoFisicoApos float64
}

// ItemSemDevolucao é um item desvinculado da OS sem movimentação de estoque.
type ItemSemDevolucao struct {
	ItemID        string
	Codigo        string
	Descricao     string
	Tipo          string
	UnidadeMedida string
	Quantidade    float64
	Motivo        string
	PedidoID      string
}

// ResultadoDevolucao é o retorno de DevolverItensAoEstoque para o caso de uso chamador.
type ResultadoDevolucao struct {
	OrdemServicoID           string
	ReservasLiberadas        []ItemLiberado
	ItensRetornadosAoEstoque []ItemRetornado
	ItensSemDevolucao        []ItemSemDevolucao
	TotalItensProcessados    int
}
