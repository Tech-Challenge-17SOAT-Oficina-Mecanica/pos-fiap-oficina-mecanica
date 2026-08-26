package mecanico

import (
	"errors"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const senhaMinima = 15

var (
	ErrNomeObrigatorio       = errors.New("nome é obrigatório")
	ErrMecanicoIDObrigatorio = errors.New("mecanicoId é obrigatório")
	ErrEmailObrigatorio      = errors.New("email é obrigatório")
	ErrSenhaObrigatoria      = errors.New("senha é obrigatória")
	ErrSenhaCurta            = errors.New("senha deve possuir no mínimo 15 caracteres")
	ErrEscoposObrigatorio    = errors.New("escopos são obrigatórios")
	ErrEmailInvalido         = errors.New("email inválido")
	ErrEscopoInvalido        = errors.New("escopo desconhecido")
)

type Mecanico struct {
	ID        string
	UsuarioID string
	Nome      string
	Email     string
	Ativo     bool
	Escopos   []string
	Version   int
}

type NovoMecanicoInput struct {
	Nome    string
	Email   string
	Senha   string
	Escopos []string
}

type AtualizarMecanicoInput struct {
	Nome    string
	Email   string
	Escopos []string
}

var escoposOficiais = map[string]struct{}{
	"mecanicos:escrever":  {},
	"clientes:ler":        {},
	"clientes:escrever":   {},
	"veiculos:ler":        {},
	"veiculos:escrever":   {},
	"os:ler":              {},
	"os:escrever":         {},
	"orcamentos:ler":      {},
	"orcamentos:escrever": {},
	"orcamentos:decidir":  {},
	"servicos:ler":        {},
	"servicos:escrever":   {},
	"estoque:ler":         {},
	"estoque:escrever":    {},
	"estoque:movimentar":  {},
	"compras:ler":         {},
	"compras:escrever":    {},
}

func Novo(input NovoMecanicoInput) (Mecanico, string, error) {
	senha := strings.TrimSpace(input.Senha)
	mecanico := Mecanico{
		Nome:    strings.TrimSpace(input.Nome),
		Email:   strings.TrimSpace(input.Email),
		Ativo:   true,
		Escopos: escoposValidos(input.Escopos),
		Version: 1,
	}
	if mecanico.Nome == "" {
		return Mecanico{}, "", ErrNomeObrigatorio
	}
	if mecanico.Email == "" {
		return Mecanico{}, "", ErrEmailObrigatorio
	}
	if !validation.IsEmail(mecanico.Email) {
		return Mecanico{}, "", ErrEmailInvalido
	}
	if senha == "" {
		return Mecanico{}, "", ErrSenhaObrigatoria
	}
	if len(senha) < senhaMinima {
		return Mecanico{}, "", ErrSenhaCurta
	}
	if len(input.Escopos) == 0 {
		return Mecanico{}, "", ErrEscoposObrigatorio
	}
	if !todosEscoposOficiais(input.Escopos) {
		return Mecanico{}, "", ErrEscopoInvalido
	}
	return mecanico, senha, nil
}

func (mecanico Mecanico) Atualizar(input AtualizarMecanicoInput) (Mecanico, error) {
	mecanico.Nome = strings.TrimSpace(input.Nome)
	mecanico.Email = strings.TrimSpace(input.Email)
	mecanico.Escopos = escoposValidos(input.Escopos)
	if mecanico.ID == "" {
		return Mecanico{}, ErrMecanicoIDObrigatorio
	}
	if mecanico.Nome == "" {
		return Mecanico{}, ErrNomeObrigatorio
	}
	if mecanico.Email == "" {
		return Mecanico{}, ErrEmailObrigatorio
	}
	if !validation.IsEmail(mecanico.Email) {
		return Mecanico{}, ErrEmailInvalido
	}
	if len(input.Escopos) == 0 {
		return Mecanico{}, ErrEscoposObrigatorio
	}
	if !todosEscoposOficiais(input.Escopos) {
		return Mecanico{}, ErrEscopoInvalido
	}
	return mecanico, nil
}

func escoposValidos(escopos []string) []string {
	seen := map[string]struct{}{}
	validos := make([]string, 0, len(escopos))
	for _, escopo := range escopos {
		escopo = strings.TrimSpace(escopo)
		if _, ok := escoposOficiais[escopo]; !ok {
			continue
		}
		if _, ok := seen[escopo]; ok {
			continue
		}
		seen[escopo] = struct{}{}
		validos = append(validos, escopo)
	}
	return validos
}

func todosEscoposOficiais(escopos []string) bool {
	for _, escopo := range escopos {
		if _, ok := escoposOficiais[strings.TrimSpace(escopo)]; !ok {
			return false
		}
	}
	return true
}
