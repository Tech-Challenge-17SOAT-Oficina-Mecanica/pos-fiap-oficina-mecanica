package seguranca

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
)

type tokenValidatorStub struct {
	claims segurancaInfrastructure.Claims
	err    error
}

func (stub tokenValidatorStub) Validar(_ string) (segurancaInfrastructure.Claims, error) {
	return stub.claims, stub.err
}

func TestRequireScopePermiteEscopo(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fornecedores", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler := RequireScope(tokenValidatorStub{claims: segurancaInfrastructure.Claims{Escopos: []string{"compras:ler"}}}, "compras:ler", http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestRequireScopeRejeitaTokenAusente(t *testing.T) {
	response := httptest.NewRecorder()
	RequireScope(tokenValidatorStub{}, "compras:ler", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/fornecedores", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestRequireScopeRejeitaEscopoInsuficiente(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fornecedores", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	RequireScope(tokenValidatorStub{claims: segurancaInfrastructure.Claims{Escopos: []string{"mecanicos:escrever"}}}, "compras:ler", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestRequireScopeRejeitaTokenInvalido(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fornecedores", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	RequireScope(tokenValidatorStub{err: errors.New("token invalido")}, "compras:ler", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", response.Code)
	}
}
