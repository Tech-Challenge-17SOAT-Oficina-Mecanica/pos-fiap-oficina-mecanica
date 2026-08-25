package orcamento_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestConsultarOrcamento(t *testing.T) {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		t.Skip("banco indisponível")
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		t.Skip("banco indisponível")
	}
	jwt, err := segurancaInfrastructure.NewJWT("segredo-de-teste")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Gerar("usuario", []string{"orcamentos:ler"})
	if err != nil {
		t.Fatal(err)
	}
	handler := presentation.NewConsultarHandler(application.NewConsultar(infrastructure.NewPostgresRepository(db)), jwt)
	cases := []struct {
		name, url, auth, contains string
		status                    int
	}{
		{"por id", "/orcamentos?orcamentoId=74000000-0000-0000-0000-000000000001", "Bearer " + token, `"valorTotalGeral":754`, http.StatusOK},
		{"principal e complementar", "/orcamentos?orcamentoId=74000000-0000-0000-0000-000000000001", "Bearer " + token, `"tipoOrcamento":"COMPLEMENTAR"`, http.StatusOK},
		{"por documento", "/orcamentos?documento=39053344705&pagina=0&tamanho=20", "Bearer " + token, `"totalElementos":2`, http.StatusOK},
		{"documento invalido", "/orcamentos?documento=11111111111", "Bearer " + token, "documento inválido", http.StatusBadRequest},
		{"sem criterio", "/orcamentos", "Bearer " + token, "informe orcamentoId ou documento", http.StatusBadRequest},
		{"nao encontrado", "/orcamentos?orcamentoId=74000000-0000-0000-0000-000000000099", "Bearer " + token, "orçamento não encontrado", http.StatusNotFound},
		{"sem token", "/orcamentos?documento=39053344705", "", "token", http.StatusUnauthorized},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.url, nil)
			if test.auth != "" {
				request.Header.Set("Authorization", test.auth)
			}
			response := httptest.NewRecorder()
			handler(response, request)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.contains) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}
