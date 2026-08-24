package cliente

import (
	"errors"
	"strings"
	"testing"
)

func TestNovoCliente(t *testing.T) {
	cases := []struct {
		name  string
		input NovoClienteInput
		want  error
	}{
		{"valido cpf telefone", NovoClienteInput{Nome: " Ana ", Documento: "39053344705", TipoDocumento: TipoDocumentoCPF, Telefone: "11988887777"}, nil},
		{"valido cnpj email", NovoClienteInput{Nome: "Empresa", Documento: "11222333000181", TipoDocumento: TipoDocumentoCNPJ, Email: "empresa@example.com"}, nil},
		{"nome ausente", NovoClienteInput{Documento: "39053344705", TipoDocumento: TipoDocumentoCPF, Telefone: "11988887777"}, ErrNomeObrigatorio},
		{"documento ausente", NovoClienteInput{Nome: "Ana", TipoDocumento: TipoDocumentoCPF, Telefone: "11988887777"}, ErrDocumentoObrigatorio},
		{"tipo ausente", NovoClienteInput{Nome: "Ana", Documento: "39053344705", Telefone: "11988887777"}, ErrTipoDocumentoObrigatorio},
		{"tipo invalido", NovoClienteInput{Nome: "Ana", Documento: "39053344705", TipoDocumento: "RG", Telefone: "11988887777"}, ErrTipoDocumentoInvalido},
		{"cpf invalido", NovoClienteInput{Nome: "Ana", Documento: "11111111111", TipoDocumento: TipoDocumentoCPF, Telefone: "11988887777"}, ErrDocumentoInvalido},
		{"cnpj invalido", NovoClienteInput{Nome: "Empresa", Documento: "11111111111111", TipoDocumento: TipoDocumentoCNPJ, Telefone: "11988887777"}, ErrDocumentoInvalido},
		{"documento com letra", NovoClienteInput{Nome: "Ana", Documento: "3905334470A", TipoDocumento: TipoDocumentoCPF, Telefone: "11988887777"}, ErrDocumentoInvalido},
		{"contato ausente", NovoClienteInput{Nome: "Ana", Documento: "39053344705", TipoDocumento: TipoDocumentoCPF}, ErrContatoObrigatorio},
		{"telefone curto", NovoClienteInput{Nome: "Ana", Documento: "39053344705", TipoDocumento: TipoDocumentoCPF, Telefone: "119"}, ErrTelefoneInvalido},
		{"telefone com letra", NovoClienteInput{Nome: "Ana", Documento: "39053344705", TipoDocumento: TipoDocumentoCPF, Telefone: "1198888777a"}, ErrTelefoneInvalido},
		{"email invalido", NovoClienteInput{Nome: "Ana", Documento: "39053344705", TipoDocumento: TipoDocumentoCPF, Email: "email"}, ErrEmailInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := Novo(test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && (!got.Ativo || got.Version != 1 || got.Nome != strings.TrimSpace(test.input.Nome)) {
				t.Fatalf("cliente: %#v", got)
			}
		})
	}
}

func TestDigitoCNPJComRestoMenorQueDois(t *testing.T) {
	if got := digitoCNPJ("00000000000000", []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}); got != 0 {
		t.Fatalf("digito: %d", got)
	}
}

func TestDocumentoParaConsulta(t *testing.T) {
	cases := []struct {
		name, documento string
		want            error
	}{
		{"cpf valido", " 39053344705 ", nil},
		{"cnpj valido", "11222333000181", nil},
		{"ausente", "", ErrDocumentoObrigatorio},
		{"letra", "3905334470A", ErrDocumentoInvalido},
		{"invalido", "11111111111", ErrDocumentoInvalido},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := DocumentoParaConsulta(test.documento)
			if !errors.Is(err, test.want) {
				t.Fatalf("erro: %v", err)
			}
			if test.want == nil && got != strings.TrimSpace(test.documento) {
				t.Fatalf("documento: %q", got)
			}
		})
	}
}
