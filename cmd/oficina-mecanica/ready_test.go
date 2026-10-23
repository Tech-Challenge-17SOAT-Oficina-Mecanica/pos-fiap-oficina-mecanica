package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type pingFunc func(context.Context) error

func (fn pingFunc) Ping(ctx context.Context) error { return fn(ctx) }

func TestReadyHandler(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"banco disponivel", nil, http.StatusOK, `{"status":"ready"}`},
		{"banco indisponivel", errors.New("credencial que nao deve vazar"), http.StatusServiceUnavailable, `{"status":"not_ready"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := pingFunc(func(ctx context.Context) error {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > time.Second {
					t.Fatal("ping precisa de timeout de no maximo um segundo")
				}
				return test.err
			})
			response := httptest.NewRecorder()
			readyHandler(db)(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
			if response.Code != test.status || response.Body.String() != test.body {
				t.Fatalf("resposta inesperada: %d %s", response.Code, response.Body.String())
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("resposta deve ser JSON")
			}
		})
	}
}

func TestReadyHandlerRespeitaCancelamento(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	db := pingFunc(func(ctx context.Context) error { return ctx.Err() })
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ready", nil).WithContext(ctx)
	readyHandler(db)(response, request)
	if response.Code != http.StatusServiceUnavailable || strings.Contains(response.Body.String(), "canceled") {
		t.Fatalf("cancelamento deve retornar indisponivel sem detalhes: %s", response.Body.String())
	}
}
