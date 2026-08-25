package fornecedor

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const prazoEntregaPadrao = 7

type Cadastro struct {
	RazaoSocial      string
	NomeFantasia     string
	Documento        string
	TipoDocumento    string
	Telefone         string
	Email            string
	PrazoEntregaDias int
}

type Fornecedor struct {
	ID string
	Cadastro
	Ativo        bool
	Version      int
	CriadoEm     time.Time
	AtualizadoEm time.Time
}

type Atualizacao struct {
	RazaoSocial      string
	NomeFantasia     string
	Telefone         string
	Email            string
	PrazoEntregaDias *int
}

func NovoCadastro(razaoSocial, nomeFantasia, documento, tipoDocumento, telefone, email string, prazoEntregaDias *int) (Cadastro, error) {
	cadastro := Cadastro{
		RazaoSocial:      strings.TrimSpace(razaoSocial),
		NomeFantasia:     strings.TrimSpace(nomeFantasia),
		Documento:        validation.OnlyDigits(documento),
		TipoDocumento:    strings.ToUpper(strings.TrimSpace(tipoDocumento)),
		Telefone:         validation.OnlyDigits(telefone),
		Email:            strings.TrimSpace(email),
		PrazoEntregaDias: prazoEntregaPadrao,
	}
	if prazoEntregaDias != nil {
		cadastro.PrazoEntregaDias = *prazoEntregaDias
	}

	if tamanhoInvalido(cadastro.RazaoSocial, 3, 120) {
		return Cadastro{}, errors.New("razaoSocial deve ter entre 3 e 120 caracteres")
	}
	if len(cadastro.NomeFantasia) > 120 {
		return Cadastro{}, errors.New("nomeFantasia deve ter no máximo 120 caracteres")
	}
	if cadastro.TipoDocumento != "CPF" && cadastro.TipoDocumento != "CNPJ" {
		return Cadastro{}, errors.New("tipoDocumento deve ser CPF ou CNPJ")
	}
	if !validation.IsDocumento(cadastro.Documento, cadastro.TipoDocumento) {
		return Cadastro{}, errors.New("documento inválido")
	}
	if cadastro.Telefone != "" && (len(cadastro.Telefone) < 10 || len(cadastro.Telefone) > 11) {
		return Cadastro{}, errors.New("telefone deve ter 10 ou 11 dígitos")
	}
	if cadastro.Email != "" && !emailValido(cadastro.Email) {
		return Cadastro{}, errors.New("email inválido")
	}
	if cadastro.Telefone == "" && cadastro.Email == "" {
		return Cadastro{}, errors.New("telefone ou email é obrigatório")
	}
	if cadastro.PrazoEntregaDias < 1 || cadastro.PrazoEntregaDias > 365 {
		return Cadastro{}, errors.New("prazoEntregaDias deve estar entre 1 e 365")
	}
	return cadastro, nil
}

func NovaAtualizacao(razaoSocial, nomeFantasia, telefone, email string, prazoEntregaDias *int) (Atualizacao, error) {
	atualizacao := Atualizacao{
		RazaoSocial:      strings.TrimSpace(razaoSocial),
		NomeFantasia:     strings.TrimSpace(nomeFantasia),
		Telefone:         validation.OnlyDigits(telefone),
		Email:            strings.TrimSpace(email),
		PrazoEntregaDias: prazoEntregaDias,
	}

	if tamanhoInvalido(atualizacao.RazaoSocial, 3, 120) {
		return Atualizacao{}, errors.New("razaoSocial deve ter entre 3 e 120 caracteres")
	}
	if len(atualizacao.NomeFantasia) > 120 {
		return Atualizacao{}, errors.New("nomeFantasia deve ter no maximo 120 caracteres")
	}
	if atualizacao.Telefone != "" && (len(atualizacao.Telefone) < 10 || len(atualizacao.Telefone) > 11) {
		return Atualizacao{}, errors.New("telefone deve ter 10 ou 11 digitos")
	}
	if atualizacao.Email != "" && !emailValido(atualizacao.Email) {
		return Atualizacao{}, errors.New("email invalido")
	}
	if atualizacao.Telefone == "" && atualizacao.Email == "" {
		return Atualizacao{}, errors.New("telefone ou email e obrigatorio")
	}
	if atualizacao.PrazoEntregaDias != nil && (*atualizacao.PrazoEntregaDias < 1 || *atualizacao.PrazoEntregaDias > 365) {
		return Atualizacao{}, errors.New("prazoEntregaDias deve estar entre 1 e 365")
	}
	return atualizacao, nil
}

func tamanhoInvalido(value string, minimo, maximo int) bool {
	size := len([]rune(value))
	return size < minimo || size > maximo
}

func emailValido(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && strings.Contains(address.Address, "@")
}
