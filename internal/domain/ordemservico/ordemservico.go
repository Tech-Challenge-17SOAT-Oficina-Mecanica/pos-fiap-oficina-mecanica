package ordemservico

import "errors"

var (
	ErrStatusNaoPermiteItens = errors.New("a ordem de servico nao permite registro de itens")
	ErrQuantidadeInvalida    = errors.New("quantidade deve ser maior que zero")
	ErrPecaFracionada        = errors.New("a quantidade de peca deve ser inteira")
)

func PermiteRegistroDeItens(status string) bool {
	switch status {
	case "RECEBIDA", "EM_DIAGNOSTICO", "AGUARDANDO_APROVACAO", "AGUARDANDO_RECURSOS", "AGUARDANDO_EXECUCAO", "EM_EXECUCAO":
		return true
	default:
		return false
	}
}

func EhComplementar(status string) bool {
	return status == "EM_EXECUCAO"
}
