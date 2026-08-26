package estoque

import "errors"

const (
	TipoPeca   = "PECA"
	TipoInsumo = "INSUMO"
)

var ErrTipoItemInvalido = errors.New("tipo do item divergente do endpoint")

func QuantidadeValida(tipo string, quantidade float64) error {
	if quantidade <= 0 {
		return errors.New("quantidade deve ser maior que zero")
	}
	if tipo == TipoPeca && quantidade != float64(int64(quantidade)) {
		return errors.New("a quantidade de peca deve ser inteira")
	}
	return nil
}
