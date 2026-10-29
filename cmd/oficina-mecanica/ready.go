package main

import (
	"context"
	"net/http"
	"time"
)

type databasePinger interface {
	Ping(context.Context) error
}

// Readiness verifica a dependencia sem expor detalhes da conexao na resposta.
// A liveness permanece independente do banco em /health.
func readyHandler(db databasePinger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), time.Second)
		defer cancel()
		writer.Header().Set("Content-Type", "application/json")
		if err := db.Ping(ctx); err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			_, _ = writer.Write([]byte(`{"status":"not_ready"}`))
			return
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(`{"status":"ready"}`))
	}
}
