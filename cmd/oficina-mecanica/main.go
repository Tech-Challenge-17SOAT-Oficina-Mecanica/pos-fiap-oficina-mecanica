package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	notificacaoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/notificacao"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	db, err := database.OpenPool()
	if err != nil {
		return err
	}
	defer db.Close()
	jwt, err := segurancaInfrastructure.NewJWT(os.Getenv("JWT_SECRET"))
	if err != nil {
		return err
	}
	life := &serverLifecycle{drainDelay: 5 * time.Second, shutdownTimeout: 20 * time.Second}
	handler, processador := montarAplicacao(db, jwt, life)
	ctx, parar := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer parar()
	server := &http.Server{Addr: ":8080", Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	log.Println("API iniciada na porta 8080")
	return life.serve(ctx, server, listener, func(workerCtx context.Context) {
		consumirNotificacoes(workerCtx, processador,
			time.Duration(inteiroDoAmbiente("NOTIFICACAO_INTERVALO_SEGUNDOS", 30))*time.Second,
			inteiroDoAmbiente("NOTIFICACAO_LOTE", 20))
	})
}

// inteiroDoAmbiente le uma configuracao numerica, caindo no padrao quando ausente ou
// invalida — a API nao deve deixar de subir por causa de uma variavel malformada.
func inteiroDoAmbiente(nome string, padrao int) int {
	valor, err := strconv.Atoi(strings.TrimSpace(os.Getenv(nome)))
	if err != nil || valor < 0 || (nome == "NOTIFICACAO_INTERVALO_SEGUNDOS" && valor == 0) {
		return padrao
	}
	return valor
}

// enviadorDeNotificacao escolhe o canal pela configuracao presente, do mais especifico
// para o mais generico. Sem nada configurado cai no log, que permite exercitar a fila
// inteira sem servidor de e-mail nenhum.
func enviadorDeNotificacao() notificacaoApplication.Enviador {
	remetente := textoDoAmbiente("EMAIL_REMETENTE", "Oficina Mecanica <onboarding@resend.dev>")

	if chave := textoDoAmbiente("RESEND_API_KEY", ""); chave != "" {
		log.Printf("notificacoes por Resend, remetente %s", remetente)
		return notificacaoInfrastructure.NewResendEnviador(chave, remetente, log.Default())
	}

	if host := textoDoAmbiente("SMTP_HOST", ""); host != "" {
		porta := inteiroDoAmbiente("SMTP_PORT", 1025)
		log.Printf("notificacoes por SMTP em %s:%d, remetente %s", host, porta, remetente)
		return notificacaoInfrastructure.NewSMTPEnviador(host, porta, remetente,
			textoDoAmbiente("SMTP_USUARIO", ""), textoDoAmbiente("SMTP_SENHA", ""))
	}

	log.Print("notificacoes apenas em log: configure SMTP_HOST ou RESEND_API_KEY para enviar")
	return notificacaoInfrastructure.NewLogEnviador(log.Default())
}

func textoDoAmbiente(nome, padrao string) string {
	if valor := strings.TrimSpace(os.Getenv(nome)); valor != "" {
		return valor
	}
	return padrao
}

// consumirNotificacoes roda ate o contexto ser cancelado. Uma rodada que falha nao
// derruba o laco: a proxima tenta de novo, e as notificacoes seguem na fila.
func consumirNotificacoes(ctx context.Context, processador notificacaoApplication.Processar, intervalo time.Duration, lote int) {
	relogio := time.NewTicker(intervalo)
	defer relogio.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-relogio.C:
			resultado, err := processador.Execute(ctx, lote)
			if err != nil {
				log.Printf("processamento de notificacoes falhou: %v", err)
				continue
			}
			if resultado.Processadas > 0 {
				log.Printf("notificacoes processadas: %d enviadas, %d falhas", resultado.Enviadas, resultado.Falhas)
			}
		}
	}
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(`{"status":"ok"}`))
}
