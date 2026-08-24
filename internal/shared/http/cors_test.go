package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusCreated)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/clientes", nil))
	if response.Code != http.StatusCreated || response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("status %d, headers %#v", response.Code, response.Header())
	}

	preflight := httptest.NewRecorder()
	handler.ServeHTTP(preflight, httptest.NewRequest(http.MethodOptions, "/clientes", nil))
	if preflight.Code != http.StatusNoContent {
		t.Fatalf("preflight status %d", preflight.Code)
	}
}
