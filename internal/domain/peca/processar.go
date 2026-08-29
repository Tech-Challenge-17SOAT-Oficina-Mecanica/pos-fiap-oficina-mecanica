package peca

type Processamento struct {
	ItemID               string
	QuantidadeSolicitada int64
	QuantidadeReservada  int64
	QuantidadeCompra     int64
	SaldoDisponivelApos  int64
}

func NovoProcessamento(itemID string, quantidade, saldoDisponivel int64) Processamento {
	reservada := quantidade
	if saldoDisponivel < reservada {
		reservada = saldoDisponivel
	}
	if reservada < 0 {
		reservada = 0
	}
	return Processamento{
		ItemID:               itemID,
		QuantidadeSolicitada: quantidade,
		QuantidadeReservada:  reservada,
		QuantidadeCompra:     quantidade - reservada,
		SaldoDisponivelApos:  saldoDisponivel - reservada,
	}
}
