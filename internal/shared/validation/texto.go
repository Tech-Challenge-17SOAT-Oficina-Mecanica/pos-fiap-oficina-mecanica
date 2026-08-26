package validation

import "strings"

var removedorAcentos = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
	"Á", "A", "À", "A", "Â", "A", "Ã", "A", "Ä", "A",
	"É", "E", "È", "E", "Ê", "E", "Ë", "E",
	"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
	"Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O", "Ö", "O",
	"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
	"Ç", "C", "Ñ", "N",
)

// NormalizarDescricao devolve a forma usada nas regras de duplicidade do catálogo:
// sem acento, sem espaço duplo e em minúsculas. Peça e insumo compartilham a regra.
func NormalizarDescricao(valor string) string {
	return strings.ToLower(strings.Join(strings.Fields(removedorAcentos.Replace(valor)), " "))
}

// ColapsarEspacos remove espaços nas pontas e reduz sequências internas a um espaço,
// preservando acentos e maiúsculas do texto original.
func ColapsarEspacos(valor string) string {
	return strings.Join(strings.Fields(valor), " ")
}
