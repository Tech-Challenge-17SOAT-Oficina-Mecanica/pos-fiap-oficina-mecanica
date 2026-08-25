package insumo

import (
	"context"
	"errors"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

type desativarRepositoryFake struct {
	item      domain.Insumo
	err       error
	insumoID  string
	usuarioID string
}

func (fake *desativarRepositoryFake) Desativar(_ context.Context, insumoID, usuarioID string) (domain.Insumo, error) {
	fake.insumoID, fake.usuarioID = insumoID, usuarioID
	return fake.item, fake.err
}

func TestDesativarInsumo(t *testing.T) {
	fake := &desativarRepositoryFake{item: domain.Insumo{ID: "insumo-1", Ativo: false}}
	item, err := NewDesativarInsumo(fake).Execute(context.Background(), " insumo-1 ", " usuario-1 ")
	if err != nil || item.ID != "insumo-1" || fake.insumoID != "insumo-1" || fake.usuarioID != "usuario-1" {
		t.Fatalf("item=%+v insumoID=%q usuarioID=%q err=%v", item, fake.insumoID, fake.usuarioID, err)
	}
}

func TestDesativarInsumoDadosInvalidos(t *testing.T) {
	for _, entrada := range [][2]string{{"", "usuario"}, {"insumo", " "}} {
		_, err := NewDesativarInsumo(&desativarRepositoryFake{}).Execute(context.Background(), entrada[0], entrada[1])
		if !errors.Is(err, ErrDesativacaoInvalida) {
			t.Fatalf("entrada=%q erro=%v", entrada, err)
		}
	}
}

func TestDesativarInsumoPropagaErro(t *testing.T) {
	esperado := errors.New("falha no banco")
	_, err := NewDesativarInsumo(&desativarRepositoryFake{err: esperado}).Execute(context.Background(), "insumo", "usuario")
	if !errors.Is(err, esperado) {
		t.Fatalf("erro=%v", err)
	}
}
