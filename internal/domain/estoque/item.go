package estoque

import (
	"errors"
	"strconv"
)

const (
	TipoPeca   = "PECA"
	TipoInsumo = "INSUMO"
)

var ErrTipoItemInvalido = errors.New("tipo do item divergente do endpoint")
var ErrQuantidadeIncompativelComUnidade = errors.New("quantidade com casas decimais incompativeis com a unidade de medida")

func QuantidadeValida(tipo string, quantidade float64) error {
	if quantidade <= 0 {
		return errors.New("quantidade deve ser maior que zero")
	}
	if tipo == TipoPeca && quantidade != float64(int64(quantidade)) {
		return errors.New("a quantidade de peca deve ser inteira")
	}
	return nil
}

// CasasDecimaisPermitidas define a precisao aceita para cada unidade de medida.
// UN nao aceita fracao; as demais seguem a precisao de NUMERIC(14,3) do schema.
func CasasDecimaisPermitidas(unidadeMedida string) int {
	if unidadeMedida == "UN" {
		return 0
	}
	return 3
}

// QuantidadeCompativelComUnidade valida se a quantidade respeita as casas decimais da unidade.
func QuantidadeCompativelComUnidade(quantidade float64, unidadeMedida string) error {
	casas := CasasDecimaisPermitidas(unidadeMedida)
	arredondada, err := strconv.ParseFloat(strconv.FormatFloat(quantidade, 'f', casas, 64), 64)
	if err != nil || arredondada != quantidade {
		return ErrQuantidadeIncompativelComUnidade
	}
	return nil
}
