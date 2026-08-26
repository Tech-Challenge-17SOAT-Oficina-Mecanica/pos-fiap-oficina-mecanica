package peca

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

const identificadorValido = "50000000-0000-0000-0000-000000000001"

type desativarFake struct {
	encontrada    peca.Peca
	erroBusca     error
	ordens        []string
	emOrcamento   bool
	erroDesativar error
	desativou     bool
	recebida      peca.Peca
}

func (fake *desativarFake) BuscarPorID(context.Context, string) (peca.Peca, error) {
	return fake.encontrada, fake.erroBusca
}

func (fake *desativarFake) OrdensComReservaAtiva(context.Context, string) ([]string, error) {
	return fake.ordens, nil
}

func (fake *desativarFake) EmOrcamentoCriado(context.Context, string) (bool, error) {
	return fake.emOrcamento, nil
}

func (fake *desativarFake) Desativar(_ context.Context, item peca.Peca) error {
	fake.desativou, fake.recebida = true, item
	return fake.erroDesativar
}

func pecaAtiva() peca.Peca {
	return peca.Peca{ID: identificadorValido, Codigo: "PEC-000001", Nome: "Filtro de oleo", Ativo: true, SaldoFisico: 9}
}

func TestDesativarPecaAtiva(t *testing.T) {
	fake := &desativarFake{encontrada: pecaAtiva()}
	useCase := NewDesativarPeca(fake)

	desativada, err := useCase.Execute(context.Background(), identificadorValido, "usuario-1")
	if err != nil {
		t.Fatal(err)
	}
	if desativada.Ativo {
		t.Fatal("peca deveria ficar inativa")
	}
	if desativada.DataDesativacao == nil || desativada.UsuarioDesativacao == nil {
		t.Fatalf("rastreabilidade nao registrada: %+v", desativada)
	}
	if *desativada.UsuarioDesativacao != "usuario-1" {
		t.Fatalf("usuarioDesativacao = %q", *desativada.UsuarioDesativacao)
	}
	if !fake.desativou {
		t.Fatal("repositorio nao foi chamado")
	}
	if fake.recebida.SaldoFisico != 9 {
		t.Fatalf("saldo fisico nao pode ser alterado, recebido %d", fake.recebida.SaldoFisico)
	}
}

func TestDesativarPecaComSaldoFisicoEPermitida(t *testing.T) {
	fake := &desativarFake{encontrada: pecaAtiva()}

	if _, err := NewDesativarPeca(fake).Execute(context.Background(), identificadorValido, "usuario-1"); err != nil {
		t.Fatalf("saldo fisico sem reserva nao deve bloquear: %v", err)
	}
}

func TestDesativarPecaJaInativa(t *testing.T) {
	inativa := pecaAtiva()
	inativa.Ativo = false
	fake := &desativarFake{encontrada: inativa}

	_, err := NewDesativarPeca(fake).Execute(context.Background(), identificadorValido, "usuario-1")
	if !errors.Is(err, peca.ErrJaInativa) {
		t.Fatalf("erro = %v, esperado ErrJaInativa", err)
	}
	if fake.desativou {
		t.Fatal("nao deveria persistir peca ja inativa")
	}
}

func TestDesativarPecaInexistente(t *testing.T) {
	fake := &desativarFake{erroBusca: ErrNaoEncontrada}

	_, err := NewDesativarPeca(fake).Execute(context.Background(), identificadorValido, "usuario-1")
	if !errors.Is(err, ErrNaoEncontrada) {
		t.Fatalf("erro = %v, esperado ErrNaoEncontrada", err)
	}
}

func TestDesativarPecaComSaldoReservado(t *testing.T) {
	fake := &desativarFake{encontrada: pecaAtiva(), ordens: []string{"os-1", "os-2"}}

	_, err := NewDesativarPeca(fake).Execute(context.Background(), identificadorValido, "usuario-1")

	var reservado ErroSaldoReservado
	if !errors.As(err, &reservado) {
		t.Fatalf("erro = %v, esperado ErroSaldoReservado", err)
	}
	if len(reservado.OrdensServico) != 2 {
		t.Fatalf("erro deve listar as OS que seguram a reserva: %+v", reservado)
	}
	if fake.desativou {
		t.Fatal("nao deveria persistir com saldo reservado")
	}
}

func TestDesativarPecaEmOrcamentoCriado(t *testing.T) {
	fake := &desativarFake{encontrada: pecaAtiva(), emOrcamento: true}

	_, err := NewDesativarPeca(fake).Execute(context.Background(), identificadorValido, "usuario-1")

	var emOrcamento ErroEmOrcamento
	if !errors.As(err, &emOrcamento) {
		t.Fatalf("erro = %v, esperado ErroEmOrcamento", err)
	}
	if fake.desativou {
		t.Fatal("nao deveria persistir com orcamento aguardando decisao")
	}
}

func TestDesativarPecaIdentificadorInvalido(t *testing.T) {
	fake := &desativarFake{encontrada: pecaAtiva()}

	_, err := NewDesativarPeca(fake).Execute(context.Background(), "abc", "usuario-1")
	if !errors.Is(err, ErrIdentificadorInvalido) {
		t.Fatalf("erro = %v, esperado ErrIdentificadorInvalido", err)
	}
	if fake.desativou {
		t.Fatal("nao deveria consultar nem persistir com identificador invalido")
	}
}

func TestDesativarRegistraMomentoDaDesativacao(t *testing.T) {
	momento := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	useCase := NewDesativarPeca(&desativarFake{encontrada: pecaAtiva()})
	useCase.agora = func() time.Time { return momento }

	desativada, err := useCase.Execute(context.Background(), identificadorValido, "usuario-1")
	if err != nil {
		t.Fatal(err)
	}
	if !desativada.DataDesativacao.Equal(momento) {
		t.Fatalf("dataDesativacao = %v, esperada %v", desativada.DataDesativacao, momento)
	}
}
