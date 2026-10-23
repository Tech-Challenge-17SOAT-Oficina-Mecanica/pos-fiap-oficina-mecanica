package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEncerramentoConcluiRequisicaoEmAndamento(t *testing.T) {
	life := &serverLifecycle{drainDelay: 50 * time.Millisecond, shutdownTimeout: time.Second}
	entered, release := make(chan struct{}), make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /lenta", func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		_, _ = w.Write([]byte("concluida"))
	})
	mux.HandleFunc("GET /ready", life.readiness(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workerStopped := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- life.serve(ctx, &http.Server{Handler: mux}, listener, func(ctx context.Context) { <-ctx.Done(); close(workerStopped) })
	}()
	response := make(chan string, 1)
	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		r, e := client.Get("http://" + listener.Addr().String() + "/lenta")
		if e != nil {
			response <- e.Error()
			return
		}
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		response <- string(b)
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("requisicao nao iniciou")
	}
	cancel()
	select {
	case <-workerStopped:
	case <-time.After(time.Second):
		t.Fatal("consumidor nao parou")
	}
	recorder := httptest.NewRecorder()
	life.readiness(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("nao deve consultar dependencia durante drenagem")
	}))(recorder, httptest.NewRequest("GET", "/ready", nil))
	if recorder.Code != 503 {
		t.Fatalf("readiness durante drenagem: %d", recorder.Code)
	}
	select {
	case err := <-done:
		t.Fatalf("servidor terminou com requisicao ativa: %v", err)
	default:
	}
	close(release)
	if body := <-response; body != "concluida" {
		t.Fatalf("requisicao interrompida: %s", body)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("servidor nao encerrou")
	}
}

func TestEncerramentoTemPrazo(t *testing.T) {
	life := &serverLifecycle{shutdownTimeout: 40 * time.Millisecond}
	entered := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done() })}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- life.serve(ctx, server, listener, func(ctx context.Context) { <-ctx.Done() }) }()
	go func() {
		client := &http.Client{Timeout: time.Second}
		r, _ := client.Get("http://" + listener.Addr().String())
		if r != nil {
			r.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("requisicao nao iniciou")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("timeout deveria ser reportado")
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown excedeu prazo")
	}
}
