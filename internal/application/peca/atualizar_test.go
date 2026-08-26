package peca

import (
	"context"
	"errors"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type atualizarRepositoryFake struct {
	retorno peca.Peca
	erro    error
	chamado bool
}

func (fake *atualizarRepositoryFake) Atualizar(_ context.Context, _ string, _ int, _ peca.Atualizacao, _ string) (peca.Peca, error) {
	fake.chamado = true
	return fake.retorno, fake.erro
}

const (
	idValido        = "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4"
	categoriaValida = "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94"
)

func TestAtualizarPecaDelegaAoRepositorio(t *testing.T) {
	fake := &atualizarRepositoryFake{retorno: peca.Peca{ID: idValido, Version: 8}}

	atualizada, err := NewAtualizarPeca(fake).Execute(context.Background(), idValido, 7,
		peca.Atualizacao{CategoriaID: categoriaValida}, "usuario-1")
	if err != nil {
		t.Fatal(err)
	}
	if !fake.chamado {
		t.Fatal("repositório não foi chamado")
	}
	if atualizada.Version != 8 {
		t.Fatalf("version = %d", atualizada.Version)
	}
}

func TestAtualizarPecaRejeitaIdentificadoresInvalidos(t *testing.T) {
	casos := []struct {
		nome      string
		id        string
		categoria string
	}{
		{"pecaId não é uuid", "nao-e-uuid", categoriaValida},
		{"categoriaId não é uuid", idValido, "nao-e-uuid"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			fake := &atualizarRepositoryFake{}
			_, err := NewAtualizarPeca(fake).Execute(context.Background(), caso.id, 7,
				peca.Atualizacao{CategoriaID: caso.categoria}, "usuario-1")
			if !errors.Is(err, ErrIdentificadorInvalido) {
				t.Fatalf("erro = %v, esperado %v", err, ErrIdentificadorInvalido)
			}
			if fake.chamado {
				t.Fatal("repositório não deveria ser chamado")
			}
		})
	}
}

func TestAtualizarPecaPropagaVersaoDivergente(t *testing.T) {
	fake := &atualizarRepositoryFake{erro: ErrVersaoDivergente}
	_, err := NewAtualizarPeca(fake).Execute(context.Background(), idValido, 7,
		peca.Atualizacao{CategoriaID: categoriaValida}, "usuario-1")
	if !errors.Is(err, ErrVersaoDivergente) {
		t.Fatalf("erro = %v", err)
	}
}
