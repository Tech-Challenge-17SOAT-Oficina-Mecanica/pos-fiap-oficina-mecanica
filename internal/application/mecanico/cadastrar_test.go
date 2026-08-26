package mecanico

import (
	"context"
	"errors"
	"strings"
	"testing"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	"golang.org/x/crypto/bcrypt"
)

type mecanicoRepositoryFake struct {
	exists          bool
	errExists       error
	saved           domain.Mecanico
	errSave         error
	senhaHash       string
	saveCalled      bool
	existsEmail     string
	existsExcept    bool
	errExistsExcept error
	found           domain.Mecanico
	errFind         error
	updated         domain.Mecanico
	errUpdate       error
	updateCalled    bool
}

func (fake *mecanicoRepositoryFake) EmailExiste(_ context.Context, email string) (bool, error) {
	fake.existsEmail = email
	return fake.exists, fake.errExists
}

func (fake *mecanicoRepositoryFake) SalvarMecanico(_ context.Context, mecanico domain.Mecanico, senhaHash string) (domain.Mecanico, error) {
	fake.saveCalled = true
	fake.senhaHash = senhaHash
	if fake.errSave != nil {
		return domain.Mecanico{}, fake.errSave
	}
	return fake.saved, nil
}

func (fake *mecanicoRepositoryFake) EmailExisteExcetoMecanico(context.Context, string, string) (bool, error) {
	return fake.existsExcept, fake.errExistsExcept
}

func (fake *mecanicoRepositoryFake) BuscarPorID(context.Context, string) (domain.Mecanico, error) {
	return fake.found, fake.errFind
}

func (fake *mecanicoRepositoryFake) AtualizarMecanico(_ context.Context, _ domain.Mecanico, _ int) (domain.Mecanico, error) {
	fake.updateCalled = true
	return fake.updated, fake.errUpdate
}

func TestCadastrarMecanico(t *testing.T) {
	input := domain.NovoMecanicoInput{Nome: "Maria", Email: "maria@oficina.local", Senha: "senha-com-15-xxx", Escopos: []string{"clientes:ler"}}
	t.Run("gera hash e salva", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{saved: domain.Mecanico{ID: "m1", Nome: "Maria", Email: "maria@oficina.local", Ativo: true, Escopos: []string{"clientes:ler"}}}
		got, err := NewCadastrar(repository).Execute(context.Background(), input)
		if err != nil || got.ID != "m1" || !repository.saveCalled {
			t.Fatalf("mecanico: %#v, erro: %v", got, err)
		}
		if repository.senhaHash == input.Senha || bcrypt.CompareHashAndPassword([]byte(repository.senhaHash), []byte(input.Senha)) != nil {
			t.Fatalf("hash invalido: %q", repository.senhaHash)
		}
	})
	t.Run("rejeita email duplicado", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{exists: true}
		_, err := NewCadastrar(repository).Execute(context.Background(), input)
		if !errors.Is(err, ErrEmailDuplicado) || repository.saveCalled {
			t.Fatalf("erro: %v, save: %v", err, repository.saveCalled)
		}
	})
	t.Run("rejeita escopo desconhecido", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{}
		_, err := NewCadastrar(repository).Execute(context.Background(), domain.NovoMecanicoInput{Nome: "Maria", Email: "maria@oficina.local", Senha: "senha-com-15-xxx", Escopos: []string{"x"}})
		if !errors.Is(err, domain.ErrEscopoInvalido) || repository.existsEmail != "" {
			t.Fatalf("erro: %v, email: %q", err, repository.existsEmail)
		}
	})
	t.Run("propaga erros", func(t *testing.T) {
		_, err := NewCadastrar(&mecanicoRepositoryFake{errExists: errors.New("db")}).Execute(context.Background(), input)
		if err == nil {
			t.Fatal("esperava erro de exists")
		}
		_, err = NewCadastrar(&mecanicoRepositoryFake{errSave: errors.New("db")}).Execute(context.Background(), input)
		if err == nil {
			t.Fatal("esperava erro de save")
		}
	})
	t.Run("propaga erro do hash", func(t *testing.T) {
		_, err := NewCadastrar(&mecanicoRepositoryFake{}).Execute(context.Background(), domain.NovoMecanicoInput{
			Nome:    "Maria",
			Email:   "maria@oficina.local",
			Senha:   strings.Repeat("a", 73),
			Escopos: []string{"clientes:ler"},
		})
		if err == nil {
			t.Fatal("esperava erro do hash")
		}
	})
}

func TestAtualizarMecanico(t *testing.T) {
	atual := domain.Mecanico{ID: "m1", UsuarioID: "u1", Nome: "Maria", Email: "maria@oficina.local", Ativo: true, Version: 2, Escopos: []string{"clientes:ler"}}
	input := AtualizarInput{MecanicoID: "m1", Version: 2, Dados: domain.AtualizarMecanicoInput{Nome: "Maria Souza", Email: "maria.souza@oficina.local", Escopos: []string{"os:ler"}}}
	t.Run("atualiza", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{found: atual, updated: domain.Mecanico{ID: "m1", Nome: "Maria Souza", Email: "maria.souza@oficina.local", Version: 3}}
		got, err := NewAtualizar(repository).Execute(context.Background(), input)
		if err != nil || got.Version != 3 || !repository.updateCalled {
			t.Fatalf("mecanico: %#v, erro: %v", got, err)
		}
	})
	t.Run("nao encontrado", func(t *testing.T) {
		_, err := NewAtualizar(&mecanicoRepositoryFake{errFind: ErrMecanicoNaoEncontrado}).Execute(context.Background(), input)
		if !errors.Is(err, ErrMecanicoNaoEncontrado) {
			t.Fatalf("erro: %v", err)
		}
	})
	t.Run("versao divergente", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{found: atual}
		_, err := NewAtualizar(repository).Execute(context.Background(), AtualizarInput{MecanicoID: "m1", Version: 1, Dados: input.Dados})
		if !errors.Is(err, ErrVersaoDivergente) || repository.updateCalled {
			t.Fatalf("erro: %v, update: %v", err, repository.updateCalled)
		}
	})
	t.Run("escopo invalido", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{found: atual}
		_, err := NewAtualizar(repository).Execute(context.Background(), AtualizarInput{MecanicoID: "m1", Version: 2, Dados: domain.AtualizarMecanicoInput{Nome: "Maria", Email: "maria@oficina.local", Escopos: []string{"x"}}})
		if !errors.Is(err, domain.ErrEscopoInvalido) || repository.updateCalled {
			t.Fatalf("erro: %v, update: %v", err, repository.updateCalled)
		}
	})
	t.Run("email duplicado", func(t *testing.T) {
		repository := &mecanicoRepositoryFake{found: atual, existsExcept: true}
		_, err := NewAtualizar(repository).Execute(context.Background(), input)
		if !errors.Is(err, ErrEmailDuplicado) || repository.updateCalled {
			t.Fatalf("erro: %v, update: %v", err, repository.updateCalled)
		}
	})
	t.Run("propaga erros", func(t *testing.T) {
		_, err := NewAtualizar(&mecanicoRepositoryFake{found: atual, errExistsExcept: errors.New("db")}).Execute(context.Background(), input)
		if err == nil {
			t.Fatal("esperava erro de exists")
		}
		_, err = NewAtualizar(&mecanicoRepositoryFake{found: atual, errUpdate: errors.New("db")}).Execute(context.Background(), input)
		if err == nil {
			t.Fatal("esperava erro de update")
		}
	})
}
