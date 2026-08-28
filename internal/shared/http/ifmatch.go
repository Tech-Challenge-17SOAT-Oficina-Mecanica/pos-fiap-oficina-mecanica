package http

import (
	"strconv"
	"strings"
)

var (
	// ErrIfMatchAusente pede 428; ErrIfMatchInvalido pede 400. Sao situacoes distintas:
	// uma e o cliente que esqueceu a pre-condicao, a outra e a pre-condicao malformada.
	ErrIfMatchAusente  = NovoErroCampo("If-Match", "If-Match é obrigatório")
	ErrIfMatchInvalido = NovoErroCampo("If-Match", "If-Match inválido")
)

// LerIfMatch extrai a version do header If-Match. Aceita tanto `3` quanto `"3"`: aspas
// sao a forma correta de ETag por HTTP, e ignora-las fazia o mesmo header ser aceito num
// modulo e rejeitado em outro.
//
// O curinga `*` nao e aceito de proposito: ele significa "atualize seja qual for a
// version", o que anularia o controle otimista que esta rota existe para garantir.
func LerIfMatch(header string) (int, error) {
	valor := strings.TrimSpace(header)
	if valor == "" {
		return 0, ErrIfMatchAusente
	}

	valor = strings.TrimSpace(strings.Trim(valor, `"`))
	version, err := strconv.Atoi(valor)
	if err != nil || version < 1 {
		return 0, ErrIfMatchInvalido
	}
	return version, nil
}
