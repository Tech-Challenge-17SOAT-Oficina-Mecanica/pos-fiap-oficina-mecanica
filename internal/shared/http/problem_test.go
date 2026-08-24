package http

import (
	"net/http/httptest"
	"testing"
)

func TestWriteProblem(t *testing.T) {
	response := httptest.NewRecorder()
	WriteProblem(response, Problem{Status: 400, Title: "Erro"})
	if response.Code != 400 || response.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatal("resposta inválida")
	}
}
