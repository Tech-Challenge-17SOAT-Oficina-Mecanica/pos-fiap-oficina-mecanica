package cliente

import (
	"errors"
	"net/mail"
	"strings"
)

const (
	TipoDocumentoCPF  = "CPF"
	TipoDocumentoCNPJ = "CNPJ"
)

var (
	ErrNomeObrigatorio          = errors.New("nome é obrigatório")
	ErrDocumentoObrigatorio     = errors.New("documento é obrigatório")
	ErrTipoDocumentoObrigatorio = errors.New("tipoDocumento é obrigatório")
	ErrTipoDocumentoInvalido    = errors.New("tipoDocumento deve ser CPF ou CNPJ")
	ErrDocumentoInvalido        = errors.New("documento inválido")
	ErrContatoObrigatorio       = errors.New("telefone ou email é obrigatório")
	ErrTelefoneInvalido         = errors.New("telefone deve ter 10 ou 11 dígitos")
	ErrEmailInvalido            = errors.New("email inválido")
	ErrClienteIDObrigatorio     = errors.New("clienteId é obrigatório")
)

type Cliente struct {
	ID            string
	Nome          string
	Documento     string
	TipoDocumento string
	Telefone      string
	Email         string
	Ativo         bool
	Version       int
	Veiculos      []Veiculo
}

type Veiculo struct {
	ID     string `json:"id"`
	Placa  string `json:"placa"`
	Marca  string `json:"marca"`
	Modelo string `json:"modelo"`
	Ano    int    `json:"ano"`
}

type NovoClienteInput struct {
	Nome          string
	Documento     string
	TipoDocumento string
	Telefone      string
	Email         string
}

type AtualizarClienteInput struct {
	Nome          string
	Documento     string
	TipoDocumento string
	Telefone      string
	Email         string
}

func Novo(input NovoClienteInput) (Cliente, error) {
	cliente := Cliente{
		Nome:          strings.TrimSpace(input.Nome),
		Documento:     strings.TrimSpace(input.Documento),
		TipoDocumento: strings.TrimSpace(input.TipoDocumento),
		Telefone:      strings.TrimSpace(input.Telefone),
		Email:         strings.TrimSpace(input.Email),
		Ativo:         true,
		Version:       1,
	}
	if err := cliente.validarCadastro(); err != nil {
		return Cliente{}, err
	}
	return cliente, nil
}

func (cliente Cliente) Atualizar(input AtualizarClienteInput) (Cliente, error) {
	cliente.Nome = strings.TrimSpace(input.Nome)
	cliente.Documento = strings.TrimSpace(input.Documento)
	cliente.TipoDocumento = strings.TrimSpace(input.TipoDocumento)
	cliente.Telefone = strings.TrimSpace(input.Telefone)
	cliente.Email = strings.TrimSpace(input.Email)
	if err := cliente.validarCadastro(); err != nil {
		return Cliente{}, err
	}
	return cliente, nil
}

func (cliente Cliente) validarCadastro() error {
	if cliente.Nome == "" {
		return ErrNomeObrigatorio
	}
	if cliente.Documento == "" {
		return ErrDocumentoObrigatorio
	}
	if cliente.TipoDocumento == "" {
		return ErrTipoDocumentoObrigatorio
	}
	if cliente.TipoDocumento != TipoDocumentoCPF && cliente.TipoDocumento != TipoDocumentoCNPJ {
		return ErrTipoDocumentoInvalido
	}
	if !somenteDigitos(cliente.Documento) || !documentoValido(cliente.Documento, cliente.TipoDocumento) {
		return ErrDocumentoInvalido
	}
	if cliente.Telefone == "" && cliente.Email == "" {
		return ErrContatoObrigatorio
	}
	if cliente.Telefone != "" && (!somenteDigitos(cliente.Telefone) || len(cliente.Telefone) < 10 || len(cliente.Telefone) > 11) {
		return ErrTelefoneInvalido
	}
	if cliente.Email != "" && !emailValido(cliente.Email) {
		return ErrEmailInvalido
	}
	return nil
}

func DocumentoParaConsulta(documento string) (string, error) {
	documento = strings.TrimSpace(documento)
	if documento == "" {
		return "", ErrDocumentoObrigatorio
	}
	if !somenteDigitos(documento) || (!cpfValido(documento) && !cnpjValido(documento)) {
		return "", ErrDocumentoInvalido
	}
	return documento, nil
}

func somenteDigitos(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != ""
}

func emailValido(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

func documentoValido(documento, tipo string) bool {
	if tipo == TipoDocumentoCPF {
		return cpfValido(documento)
	}
	return cnpjValido(documento)
}

func cpfValido(cpf string) bool {
	if len(cpf) != 11 || todosIguais(cpf) {
		return false
	}
	return digitoCPF(cpf, 9) == int(cpf[9]-'0') && digitoCPF(cpf, 10) == int(cpf[10]-'0')
}

func digitoCPF(cpf string, pos int) int {
	sum := 0
	for i := 0; i < pos; i++ {
		sum += int(cpf[i]-'0') * (pos + 1 - i)
	}
	digit := (sum * 10) % 11
	if digit == 10 {
		return 0
	}
	return digit
}

func cnpjValido(cnpj string) bool {
	if len(cnpj) != 14 || todosIguais(cnpj) {
		return false
	}
	return digitoCNPJ(cnpj, []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}) == int(cnpj[12]-'0') &&
		digitoCNPJ(cnpj, []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}) == int(cnpj[13]-'0')
}

func digitoCNPJ(cnpj string, weights []int) int {
	sum := 0
	for i, weight := range weights {
		sum += int(cnpj[i]-'0') * weight
	}
	rest := sum % 11
	if rest < 2 {
		return 0
	}
	return 11 - rest
}

func todosIguais(value string) bool {
	for i := 1; i < len(value); i++ {
		if value[i] != value[0] {
			return false
		}
	}
	return true
}
