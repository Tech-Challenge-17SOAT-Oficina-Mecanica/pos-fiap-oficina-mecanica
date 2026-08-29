package notificacao

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

type filaFake struct {
	pendentes   []notificacao.Notificacao
	erroLeitura error
	atualizadas []notificacao.Notificacao
}

func (fake *filaFake) Pendentes(_ context.Context, limite int) ([]notificacao.Notificacao, error) {
	if fake.erroLeitura != nil {
		return nil, fake.erroLeitura
	}
	if limite < len(fake.pendentes) {
		return fake.pendentes[:limite], nil
	}
	return fake.pendentes, nil
}

func (fake *filaFake) AtualizarResultado(_ context.Context, aviso notificacao.Notificacao) error {
	fake.atualizadas = append(fake.atualizadas, aviso)
	return nil
}

type enviadorFake struct {
	falharEm map[string]error
	enviadas []string
}

func (fake *enviadorFake) Enviar(_ context.Context, aviso notificacao.Notificacao) error {
	if err, existe := fake.falharEm[aviso.ID]; existe {
		return err
	}
	fake.enviadas = append(fake.enviadas, aviso.ID)
	return nil
}

func fila(ids ...string) []notificacao.Notificacao {
	var avisos []notificacao.Notificacao
	for _, id := range ids {
		avisos = append(avisos, notificacao.Notificacao{ID: id, Status: notificacao.StatusPendente})
	}
	return avisos
}

func TestProcessarEnviaEMarca(t *testing.T) {
	repo := &filaFake{pendentes: fila("a", "b")}
	enviador := &enviadorFake{}

	resultado, err := NewProcessar(repo, enviador).Execute(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Enviadas != 2 || resultado.Falhas != 0 {
		t.Fatalf("resultado = %+v", resultado)
	}
	for _, atualizada := range repo.atualizadas {
		if atualizada.Status != notificacao.StatusEnviada {
			t.Fatalf("status = %q, esperado ENVIADA", atualizada.Status)
		}
		if atualizada.EnviadaEm == nil {
			t.Fatal("a data de envio deveria ter sido gravada")
		}
	}
}

// Uma falha individual não pode interromper a fila: as demais precisam seguir.
func TestProcessarIsolaFalhaIndividual(t *testing.T) {
	repo := &filaFake{pendentes: fila("a", "quebrada", "c")}
	enviador := &enviadorFake{falharEm: map[string]error{"quebrada": errors.New("smtp fora do ar")}}

	resultado, err := NewProcessar(repo, enviador).Execute(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Processadas != 3 || resultado.Enviadas != 2 || resultado.Falhas != 1 {
		t.Fatalf("resultado = %+v; a falha do meio não podia parar a fila", resultado)
	}

	var falhou notificacao.Notificacao
	for _, atualizada := range repo.atualizadas {
		if atualizada.ID == "quebrada" {
			falhou = atualizada
		}
	}
	if falhou.Status != notificacao.StatusFalhou {
		t.Fatalf("status = %q, esperado FALHOU", falhou.Status)
	}
	if falhou.UltimoErro == nil || *falhou.UltimoErro != "smtp fora do ar" {
		t.Fatal("o motivo da falha deveria ter sido gravado para diagnóstico")
	}
	if !falhou.Reenviavel() {
		t.Fatal("a que falhou precisa voltar na próxima rodada")
	}
}

func TestProcessarRespeitaOLimite(t *testing.T) {
	repo := &filaFake{pendentes: fila("a", "b", "c", "d")}

	resultado, err := NewProcessar(repo, &enviadorFake{}).Execute(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Processadas != 2 {
		t.Fatalf("processadas = %d, esperado 2", resultado.Processadas)
	}
}

func TestProcessarFilaVazia(t *testing.T) {
	resultado, err := NewProcessar(&filaFake{}, &enviadorFake{}).Execute(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Processadas != 0 {
		t.Fatalf("resultado = %+v", resultado)
	}
}

func TestProcessarPropagaFalhaDeLeitura(t *testing.T) {
	repo := &filaFake{erroLeitura: context.DeadlineExceeded}

	if _, err := NewProcessar(repo, &enviadorFake{}).Execute(context.Background(), 10); err == nil {
		t.Fatal("a falha ao ler a fila deveria ser propagada")
	}
	if len(repo.atualizadas) != 0 {
		t.Fatal("nada podia ter sido atualizado")
	}
}

func TestProcessarRetentaOQueFalhou(t *testing.T) {
	anterior := "tentativa anterior"
	repo := &filaFake{pendentes: []notificacao.Notificacao{
		{ID: "a", Status: notificacao.StatusFalhou, Tentativas: 1, UltimoErro: &anterior},
	}}

	resultado, err := NewProcessar(repo, &enviadorFake{}).Execute(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if resultado.Enviadas != 1 {
		t.Fatalf("resultado = %+v; o que falhou precisa ser retentado", resultado)
	}
	entregue := repo.atualizadas[0]
	if entregue.Tentativas != 2 {
		t.Fatalf("tentativas = %d, esperado 2", entregue.Tentativas)
	}
	if entregue.UltimoErro != nil {
		t.Fatal("o erro anterior deveria ser limpo após o sucesso")
	}
}
