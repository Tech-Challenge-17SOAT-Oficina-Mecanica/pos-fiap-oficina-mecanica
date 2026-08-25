package seguranca

import (
	"context"
	"net/http"
	"slices"
	"strings"

	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type Autenticador interface {
	Autenticar(token string) (usuarioID string, escopos []string, err error)
}

type chaveContexto string

const chaveUsuarioID chaveContexto = "usuarioID"

func UsuarioID(ctx context.Context) string {
	usuarioID, _ := ctx.Value(chaveUsuarioID).(string)
	return usuarioID
}

func ComEscopo(autenticador Autenticador, escopo string, proximo http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		token, ok := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			problemaAutorizacao(writer, http.StatusUnauthorized, "Não autorizado", "token ausente")
			return
		}

		usuarioID, escopos, err := autenticador.Autenticar(strings.TrimSpace(token))
		if err != nil {
			problemaAutorizacao(writer, http.StatusUnauthorized, "Não autorizado", "token inválido ou expirado")
			return
		}
		if !slices.Contains(escopos, escopo) {
			problemaAutorizacao(writer, http.StatusForbidden, "Acesso negado", "escopo "+escopo+" é obrigatório")
			return
		}

		proximo(writer, request.WithContext(context.WithValue(request.Context(), chaveUsuarioID, usuarioID)))
	}
}

func problemaAutorizacao(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{
		Type:   "https://api.oficina-mecanica.dev/errors/autorizacao",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}
