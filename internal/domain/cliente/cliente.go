package cliente

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
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
	ErrMotivoInvalido           = errors.New("motivo deve ter no máximo 200 caracteres")
)

type Cliente struct {
	ID            string
	Nome          string
	Documento     string
	TipoDocumento string
	Telefone      string
	Email         string
	Ativo         bool
	InativadoEm   *time.Time
	InativadoPor  string
	Motivo        string
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

type VeiculoInativado struct {
	ID    string `json:"id"`
	Placa string `json:"placa"`
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

func MotivoParaInativacao(motivo string) (string, error) {
	motivo = strings.TrimSpace(motivo)
	if len(motivo) > 200 {
		return "", ErrMotivoInvalido
	}
	return motivo, nil
}

func Novo(input NovoClienteInput) (Cliente, error) {
	cliente := Cliente{
		Nome:          strings.TrimSpace(input.Nome),
		Documento:     validation.OnlyDigits(input.Documento),
		TipoDocumento: strings.ToUpper(strings.TrimSpace(input.TipoDocumento)),
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
	cliente.Documento = validation.OnlyDigits(input.Documento)
	cliente.TipoDocumento = strings.ToUpper(strings.TrimSpace(input.TipoDocumento))
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
	if !validation.IsDocumento(cliente.Documento, cliente.TipoDocumento) {
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
	documento = validation.OnlyDigits(documento)
	if documento == "" {
		return "", ErrDocumentoObrigatorio
	}
	if !validation.IsCPF(documento) && !validation.IsCNPJ(documento) {
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
