package seguranca

import (
	"context"
	"net/http"
	"strings"

	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type claimsContextKey struct{}

type TokenValidator interface {
	Validar(string) (segurancaInfrastructure.Claims, error)
}

func RequireScope(tokenValidator TokenValidator, escopo string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		raw := strings.TrimSpace(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "))
		if raw == "" {
			sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/autenticacao", Title: "Nao autenticado", Status: http.StatusUnauthorized, Detail: "token ausente"})
			return
		}
		claims, err := tokenValidator.Validar(raw)
		if err != nil {
			sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/autenticacao", Title: "Nao autenticado", Status: http.StatusUnauthorized, Detail: "token invalido"})
			return
		}
		if !temEscopo(claims.Escopos, escopo) {
			sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/autorizacao", Title: "Acesso negado", Status: http.StatusForbidden, Detail: "escopo insuficiente"})
			return
		}
		request = request.WithContext(context.WithValue(request.Context(), claimsContextKey{}, claims))
		next.ServeHTTP(writer, request)
	})
}

func ClaimsFromContext(ctx context.Context) (segurancaInfrastructure.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(segurancaInfrastructure.Claims)
	return claims, ok
}

func temEscopo(escopos []string, esperado string) bool {
	for _, escopo := range escopos {
		if escopo == esperado {
			return true
		}
	}
	return false
}
