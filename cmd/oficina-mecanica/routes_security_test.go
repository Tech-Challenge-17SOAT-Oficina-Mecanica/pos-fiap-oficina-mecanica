package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
)

func TestRotasProtegidasValidamAutenticacaoEPermissao(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("SMTP_HOST", "")
	jwt, err := segurancaInfrastructure.NewJWT("chave-exclusiva-teste")
	if err != nil {
		t.Fatal(err)
	}
	semEscopo, err := jwt.Gerar("90000000-0000-0000-0000-000000000001", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler, _ := montarAplicacao(nil, jwt, &serverLifecycle{})
	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`mux\.Handle\("([A-Z]+) ([^"]+)"`)
	parameter := regexp.MustCompile(`\{[^}]+\}`)
	for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
		method, path := match[1], match[2]
		if path == "/autenticacao/login" {
			continue
		}
		path = parameter.ReplaceAllString(path, "10000000-0000-0000-0000-000000000001")
		t.Run(method+" "+path, func(t *testing.T) {
			for _, scenario := range []struct {
				name, authorization string
				status              int
			}{
				{"sem token", "", http.StatusUnauthorized},
				{"token invalido", "Bearer invalido", http.StatusUnauthorized},
				{"sem permissao", "Bearer " + semEscopo, http.StatusForbidden},
			} {
				t.Run(scenario.name, func(t *testing.T) {
					request := httptest.NewRequest(method, path, strings.NewReader(`{}`))
					request.Header.Set("Authorization", scenario.authorization)
					response := httptest.NewRecorder()
					handler.ServeHTTP(response, request)
					if response.Code != scenario.status {
						t.Fatalf("status=%d esperado=%d body=%s", response.Code, scenario.status, response.Body.String())
					}
					if response.Header().Get("Content-Type") != "application/problem+json" {
						t.Fatalf("tipo de resposta inesperado: %s", response.Header().Get("Content-Type"))
					}
				})
			}
		})
	}
}

func TestMontagemDaAplicacaoMantemSaudeECORS(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("SMTP_HOST", "")
	jwt, err := segurancaInfrastructure.NewJWT("chave-exclusiva-teste")
	if err != nil {
		t.Fatal(err)
	}
	handler, _ := montarAplicacao(nil, jwt, &serverLifecycle{})
	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK || health.Body.String() != `{"status":"ok"}` {
		t.Fatalf("health=%d %s", health.Code, health.Body.String())
	}
	preflight := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/clientes", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	handler.ServeHTTP(preflight, request)
	if preflight.Code != http.StatusNoContent || preflight.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("preflight=%d headers=%v", preflight.Code, preflight.Header())
	}
}
