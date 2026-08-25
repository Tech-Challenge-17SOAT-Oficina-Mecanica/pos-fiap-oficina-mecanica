package seguranca

import (
	"net/http"
	"strings"

	security "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type tokenValidator interface {
	Validar(string) (security.Claims, error)
}

func RequireScope(token tokenValidator, scope string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		parts := strings.Fields(request.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(writer)
			return
		}
		claims, err := token.Validar(parts[1])
		if err != nil {
			unauthorized(writer)
			return
		}
		for _, granted := range claims.Escopos {
			if granted == scope {
				next.ServeHTTP(writer, request)
				return
			}
		}
		sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/autorizacao", Title: "Acesso negado", Status: http.StatusForbidden, Detail: "usuário sem permissão para esta operação"})
	})
}

func unauthorized(writer http.ResponseWriter) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/autenticacao", Title: "Não autorizado", Status: http.StatusUnauthorized, Detail: "token de acesso inválido ou ausente"})
}
