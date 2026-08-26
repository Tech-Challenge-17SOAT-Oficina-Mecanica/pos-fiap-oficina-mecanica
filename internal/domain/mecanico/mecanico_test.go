package mecanico

import (
	"errors"
	"testing"
)

func TestNovoMecanico(t *testing.T) {
	cases := []struct {
		name  string
		input NovoMecanicoInput
		want  error
	}{
		{"nome obrigatorio", NovoMecanicoInput{Email: "m@oficina.local", Senha: "senha-com-15-xxx", Escopos: []string{"clientes:ler"}}, ErrNomeObrigatorio},
		{"email obrigatorio", NovoMecanicoInput{Nome: "Maria", Senha: "senha-com-15-xxx", Escopos: []string{"clientes:ler"}}, ErrEmailObrigatorio},
		{"email invalido", NovoMecanicoInput{Nome: "Maria", Email: "maria", Senha: "senha-com-15-xxx", Escopos: []string{"clientes:ler"}}, ErrEmailInvalido},
		{"senha obrigatoria", NovoMecanicoInput{Nome: "Maria", Email: "m@oficina.local", Escopos: []string{"clientes:ler"}}, ErrSenhaObrigatoria},
		{"senha curta", NovoMecanicoInput{Nome: "Maria", Email: "m@oficina.local", Senha: "senha-curta", Escopos: []string{"clientes:ler"}}, ErrSenhaCurta},
		{"escopos obrigatorios", NovoMecanicoInput{Nome: "Maria", Email: "m@oficina.local", Senha: "senha-com-15-xxx"}, ErrEscoposObrigatorio},
		{"escopo desconhecido", NovoMecanicoInput{Nome: "Maria", Email: "m@oficina.local", Senha: "senha-com-15-xxx", Escopos: []string{"invalido"}}, ErrEscopoInvalido},
		{"sucesso", NovoMecanicoInput{Nome: " Maria ", Email: "m@oficina.local", Senha: " senha-com-15-xxx ", Escopos: []string{" clientes:ler ", "clientes:ler", "os:ler"}}, nil},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, senha, err := Novo(test.input)
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || got.Nome != "Maria" || senha != "senha-com-15-xxx" || len(got.Escopos) != 2 || !got.Ativo || got.Version != 1) {
				t.Fatalf("mecanico: %#v, senha: %q, erro: %v", got, senha, err)
			}
		})
	}
}

func TestAtualizarMecanico(t *testing.T) {
	base := Mecanico{ID: "m1", UsuarioID: "u1", Version: 1, Ativo: true}
	cases := []struct {
		name  string
		base  Mecanico
		input AtualizarMecanicoInput
		want  error
	}{
		{"id obrigatorio", Mecanico{}, AtualizarMecanicoInput{Nome: "Maria", Email: "m@oficina.local", Escopos: []string{"clientes:ler"}}, ErrMecanicoIDObrigatorio},
		{"nome obrigatorio", base, AtualizarMecanicoInput{Email: "m@oficina.local", Escopos: []string{"clientes:ler"}}, ErrNomeObrigatorio},
		{"email obrigatorio", base, AtualizarMecanicoInput{Nome: "Maria", Escopos: []string{"clientes:ler"}}, ErrEmailObrigatorio},
		{"email invalido", base, AtualizarMecanicoInput{Nome: "Maria", Email: "maria", Escopos: []string{"clientes:ler"}}, ErrEmailInvalido},
		{"escopos obrigatorios", base, AtualizarMecanicoInput{Nome: "Maria", Email: "m@oficina.local"}, ErrEscoposObrigatorio},
		{"escopo desconhecido", base, AtualizarMecanicoInput{Nome: "Maria", Email: "m@oficina.local", Escopos: []string{"x"}}, ErrEscopoInvalido},
		{"sucesso", base, AtualizarMecanicoInput{Nome: " Maria ", Email: "m@oficina.local", Escopos: []string{" os:ler ", "os:ler"}}, nil},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.base.Atualizar(test.input)
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (err != nil || got.Nome != "Maria" || len(got.Escopos) != 1 || got.Version != 1) {
				t.Fatalf("mecanico: %#v, erro: %v", got, err)
			}
		})
	}
}
