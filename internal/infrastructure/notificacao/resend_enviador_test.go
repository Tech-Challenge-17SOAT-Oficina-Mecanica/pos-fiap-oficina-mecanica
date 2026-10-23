package notificacao

import (
	"context"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestResendReutilizaChaveIdempotencia(t *testing.T) {
	var keys []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"email-teste"}`)
	}))
	defer server.Close()
	sender := NewResendEnviador("teste", "oficina@example.com", log.New(io.Discard, "", 0))
	sender.cliente.BaseURL, _ = url.Parse(server.URL + "/")
	aviso := domain.Notificacao{ID: "id-estavel", Destinatario: "teste@example.com", Assunto: "Teste", Conteudo: "Teste"}
	for i := 0; i < 2; i++ {
		if err := sender.Enviar(context.Background(), aviso); err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 2 || keys[0] != "notificacao/id-estavel" || keys[1] != keys[0] {
		t.Fatalf("chaves=%v", keys)
	}
}
