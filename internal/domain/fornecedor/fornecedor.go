package fornecedor

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode"
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
	InativadoEm  *time.Time
	InativadoPor string
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
		Documento:        apenasDigitos(documento),
		TipoDocumento:    strings.ToUpper(strings.TrimSpace(tipoDocumento)),
		Telefone:         apenasDigitos(telefone),
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
		return Cadastro{}, errors.New("nomeFantasia deve ter no maximo 120 caracteres")
	}
	if cadastro.TipoDocumento != "CPF" && cadastro.TipoDocumento != "CNPJ" {
		return Cadastro{}, errors.New("tipoDocumento deve ser CPF ou CNPJ")
	}
	if !documentoValido(cadastro.Documento, cadastro.TipoDocumento) {
		return Cadastro{}, errors.New("documento invalido")
	}
	if cadastro.Telefone != "" && (len(cadastro.Telefone) < 10 || len(cadastro.Telefone) > 11) {
		return Cadastro{}, errors.New("telefone deve ter 10 ou 11 digitos")
	}
	if cadastro.Email != "" && !emailValido(cadastro.Email) {
		return Cadastro{}, errors.New("email invalido")
	}
	if cadastro.Telefone == "" && cadastro.Email == "" {
		return Cadastro{}, errors.New("telefone ou email e obrigatorio")
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
		Telefone:         apenasDigitos(telefone),
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

func apenasDigitos(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsDigit(character) {
			return character
		}
		return -1
	}, value)
}

func emailValido(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && strings.Contains(address.Address, "@")
}

func documentoValido(documento, tipo string) bool {
	if tipo == "CPF" {
		return cpfValido(documento)
	}
	return cnpjValido(documento)
}

func cpfValido(value string) bool {
	if len(value) != 11 || todosDigitosIguais(value) {
		return false
	}
	return digitoVerificador(value[:9], 10) == int(value[9]-'0') && digitoVerificador(value[:10], 11) == int(value[10]-'0')
}

func cnpjValido(value string) bool {
	if len(value) != 14 || todosDigitosIguais(value) {
		return false
	}
	weightsFirst := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	weightsSecond := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	return digitoComPesos(value[:12], weightsFirst) == int(value[12]-'0') && digitoComPesos(value[:13], weightsSecond) == int(value[13]-'0')
}

func todosDigitosIguais(value string) bool {
	for index := 1; index < len(value); index++ {
		if value[index] != value[0] {
			return false
		}
	}
	return true
}

func digitoVerificador(value string, weight int) int {
	sum := 0
	for _, digit := range value {
		sum += int(digit-'0') * weight
		weight--
	}
	remainder := (sum * 10) % 11
	if remainder == 10 {
		return 0
	}
	return remainder
}

func digitoComPesos(value string, weights []int) int {
	sum := 0
	for index, digit := range value {
		sum += int(digit-'0') * weights[index]
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
