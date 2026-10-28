package notificacao

import (
	"context"
	"testing"
	"time"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/notificacao"
)

func TestPostgresRepositoryNotificacao(t *testing.T) {
	ctx := context.Background()
	db := bancoIsolado(t)
	var err error

	const clienteID = "20000000-0000-0000-1000-000000000001"
	const agregadoID = "70000000-0000-0000-1000-000000000001"
	_, _ = db.Exec(ctx, "DELETE FROM notificacao WHERE cliente_id = $1 OR agregado_id = $2", clienteID, agregadoID)
	_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id = $1", clienteID)
	defer db.Exec(ctx, "DELETE FROM notificacao WHERE cliente_id = $1 OR agregado_id = $2", clienteID, agregadoID)
	defer db.Exec(ctx, "DELETE FROM cliente WHERE id = $1", clienteID)

	if _, err = db.Exec(ctx, `
		INSERT INTO cliente (id, nome, documento, tipo_documento, telefone, email, ativo)
		VALUES ($1, 'Cliente Notificacao', '12345678909', 'CPF', '11999999999', 'cliente@example.com', TRUE)`,
		clienteID); err != nil {
		t.Fatal(err)
	}

	repository := infrastructure.NewPostgresRepository(db)
	contato, err := repository.ContatoDoCliente(ctx, clienteID)
	if err != nil || contato.Email != "cliente@example.com" {
		t.Fatalf("contato=%+v err=%v", contato, err)
	}

	aviso, err := domain.NovaComHTML(clienteID, "destino@example.com", domain.EventoServicoFinalizado,
		"Servico finalizado", "Retire o veiculo.", "<p>Retire o veiculo.</p>",
		domain.Origem{Agregado: "ordem_servico", ID: agregadoID})
	if err != nil {
		t.Fatal(err)
	}
	aviso, err = repository.Enfileirar(ctx, aviso)
	if err != nil || aviso.ID == "" || aviso.CriadaEm.IsZero() {
		t.Fatalf("aviso=%+v err=%v", aviso, err)
	}

	_, err = db.Exec(ctx, "UPDATE notificacao SET criada_em='2000-01-01' WHERE id=$1", aviso.ID)
	if err != nil {
		t.Fatal(err)
	}
	err = repository.ProcessarPendentes(ctx, 1, func(_ context.Context, atual domain.Notificacao) (domain.Notificacao, error) {
		if atual.ID != aviso.ID || atual.ConteudoHTML == "" {
			t.Errorf("notificacao inesperada: %+v", atual)
		}
		return atual.MarcarEnviada(time.Now().UTC())
	})
	if err != nil {
		t.Fatal(err)
	}
	var status string
	var tentativas int
	if err = db.QueryRow(ctx, "SELECT status, tentativas FROM notificacao WHERE id = $1", aviso.ID).Scan(&status, &tentativas); err != nil {
		t.Fatal(err)
	}
	if status != domain.StatusEnviada || tentativas != 1 {
		t.Fatalf("status=%s tentativas=%d", status, tentativas)
	}
}
