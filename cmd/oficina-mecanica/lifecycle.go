package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// O periodo de drenagem permite propagar a retirada do endpoint do Service.
// Drenagem (5s) + shutdown (20s) ficam abaixo dos 30s de graca do pod.
type serverLifecycle struct {
	draining        atomic.Bool
	drainDelay      time.Duration
	shutdownTimeout time.Duration
}

func (life *serverLifecycle) readiness(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if life.draining.Load() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"not_ready"}`))
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (life *serverLifecycle) serve(ctx context.Context, server *http.Server, listener net.Listener, worker func(context.Context)) error {
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	workerDone := make(chan struct{})
	go func() { defer close(workerDone); worker(workerCtx) }()
	serverDone := make(chan error, 1)
	go func() { serverDone <- server.Serve(listener) }()
	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-serverDone:
	}
	life.draining.Store(true)
	cancelWorker()
	timer := time.NewTimer(life.drainDelay)
	<-timer.C
	shutdownCtx, cancel := context.WithTimeout(context.Background(), life.shutdownTimeout)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		_ = server.Close()
	}
	select {
	case <-workerDone:
	case <-shutdownCtx.Done():
		shutdownErr = errors.Join(shutdownErr, shutdownCtx.Err())
	}
	if serveErr == nil {
		serveErr = <-serverDone
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(serveErr, shutdownErr)
}
