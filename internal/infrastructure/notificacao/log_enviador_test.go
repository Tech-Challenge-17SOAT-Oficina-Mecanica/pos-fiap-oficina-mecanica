package notificacao

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

func TestLogEnviadorNaoRegistraConteudoDaMensagem(t *testing.T) {
	var buffer bytes.Buffer
	enviador := NewLogEnviador(log.New(&buffer, "", 0))
	aviso := domain.Notificacao{
		ID:         "notificacao-id",
		Canal:      domain.CanalEmail,
		TipoEvento: domain.EventoOrcamentoPronto,
		Origem: domain.Origem{
			Agregado: "ORCAMENTO",
			ID:       "orcamento-id",
		},
		Assunto:  "Orcamento pronto",
		Conteudo: "conteudo sensivel do cliente",
	}

	if err := enviador.Enviar(context.Background(), aviso); err != nil {
		t.Fatal(err)
	}

	logado := buffer.String()
	for _, esperado := range []string{"notificacao-id", "canal=EMAIL", "evento=ORCAMENTO_PRONTO", "origem=ORCAMENTO/orcamento-id", `assunto="Orcamento pronto"`} {
		if !strings.Contains(logado, esperado) {
			t.Fatalf("log nao contem %q: %s", esperado, logado)
		}
	}
	if strings.Contains(logado, aviso.Conteudo) {
		t.Fatalf("conteudo sensivel apareceu no log: %s", logado)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if NewPostgresRepository(&pgxpool.Pool{}).db == nil {
		t.Fatal("db obrigatorio")
	}
}

type linhaNotificacaoFake struct {
	err error
}

func (linha linhaNotificacaoFake) Scan(destinos ...any) error {
	if linha.err != nil {
		return linha.err
	}
	ultimoErro := "smtp"
	enviadaEm := time.Now()
	*(destinos[0].(*string)) = "notificacao-id"
	*(destinos[1].(*string)) = "cliente-id"
	*(destinos[2].(*string)) = domain.CanalEmail
	*(destinos[3].(*string)) = domain.EventoOrcamentoPronto
	*(destinos[4].(*string)) = "ORCAMENTO"
	*(destinos[5].(*string)) = "orcamento-id"
	*(destinos[6].(*string)) = "cliente@example.com"
	*(destinos[7].(*string)) = "Orcamento pronto"
	*(destinos[8].(*string)) = "Conteudo"
	*(destinos[9].(*string)) = domain.StatusFalhou
	*(destinos[10].(*int)) = 2
	*(destinos[11].(**string)) = &ultimoErro
	*(destinos[12].(*time.Time)) = time.Now()
	*(destinos[13].(**time.Time)) = &enviadaEm
	return nil
}

func TestLer(t *testing.T) {
	aviso, err := ler(linhaNotificacaoFake{})
	if err != nil {
		t.Fatal(err)
	}
	if aviso.ID != "notificacao-id" || aviso.Origem.Agregado != "ORCAMENTO" || aviso.UltimoErro == nil || aviso.EnviadaEm == nil {
		t.Fatalf("notificacao invalida: %+v", aviso)
	}

	if _, err := ler(linhaNotificacaoFake{err: errors.New("db")}); err == nil {
		t.Fatal("esperava erro")
	}
}
