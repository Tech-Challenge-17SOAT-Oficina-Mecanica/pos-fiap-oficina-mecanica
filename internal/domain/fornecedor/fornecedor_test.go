package fornecedor

import (
	"strings"
	"testing"
)

func TestNovoCadastroNormalizaDocumentoEAplicaPrazoPadrao(t *testing.T) {
	cadastro, err := NovoCadastro(
		"Auto Pecas Brasil Ltda",
		"Auto Pecas",
		"04.252.011/0001-10",
		"cnpj",
		"(11) 99999-0000",
		"",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if cadastro.Documento != "04252011000110" {
		t.Fatalf("documento=%q", cadastro.Documento)
	}
	if cadastro.PrazoEntregaDias != 7 {
		t.Fatalf("prazo=%d", cadastro.PrazoEntregaDias)
	}
	if cadastro.Telefone != "11999990000" {
		t.Fatalf("telefone=%q", cadastro.Telefone)
	}
}

func TestNovoCadastroRejeitaDadosInvalidos(t *testing.T) {
	tests := []struct {
		name      string
		documento string
		tipo      string
		telefone  string
		email     string
	}{
		{name: "documento invalido", documento: "123", tipo: "CNPJ", telefone: "11999990000"},
		{name: "sem contato", documento: "52998224725", tipo: "CPF"},
		{name: "telefone invalido", documento: "52998224725", tipo: "CPF", telefone: "123"},
		{name: "email invalido", documento: "52998224725", tipo: "CPF", email: "invalido"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NovoCadastro("Fornecedor valido", "", test.documento, test.tipo, test.telefone, test.email, nil); err == nil {
				t.Fatal("cadastro invalido aceito")
			}
		})
	}
}

func TestNovoCadastroRejeitaDemaisLimites(t *testing.T) {
	prazoInvalido := 0
	for _, test := range []struct {
		razao, fantasia, tipo string
		prazo                 *int
	}{
		{"ab", "", "CNPJ", nil},
		{"Fornecedor válido", strings.Repeat("a", 121), "CNPJ", nil},
		{"Fornecedor válido", "", "RG", nil},
		{"Fornecedor válido", "", "CNPJ", &prazoInvalido},
		{"Fornecedor válido", "", "CPF", nil},
	} {
		if _, err := NovoCadastro(test.razao, test.fantasia, "04.252.011/0001-10", test.tipo, "11999990000", "", test.prazo); err == nil {
			t.Fatal("cadastro inválido aceito")
		}
	}
}

func TestNovaAtualizacaoNormalizaCampos(t *testing.T) {
	prazo := 15
	atualizacao, err := NovaAtualizacao(" Fornecedor Atualizado ", " Fantasia ", "(11) 98888-7777", "contato@example.com", &prazo)
	if err != nil {
		t.Fatal(err)
	}
	if atualizacao.RazaoSocial != "Fornecedor Atualizado" {
		t.Fatalf("razaoSocial=%q", atualizacao.RazaoSocial)
	}
	if atualizacao.NomeFantasia != "Fantasia" || atualizacao.Telefone != "11988887777" || atualizacao.Email != "contato@example.com" {
		t.Fatalf("atualizacao invalida: %+v", atualizacao)
	}
	if atualizacao.PrazoEntregaDias == nil || *atualizacao.PrazoEntregaDias != prazo {
		t.Fatalf("prazo=%v", atualizacao.PrazoEntregaDias)
	}
}

func TestNovaAtualizacaoRejeitaDadosInvalidos(t *testing.T) {
	prazoZero := 0
	prazoAlto := 366
	for _, test := range []struct {
		nome     string
		razao    string
		fantasia string
		telefone string
		email    string
		prazo    *int
	}{
		{"razao curta", "ab", "", "11999990000", "", nil},
		{"fantasia longa", "Fornecedor valido", strings.Repeat("a", 121), "11999990000", "", nil},
		{"telefone invalido", "Fornecedor valido", "", "123", "", nil},
		{"email invalido", "Fornecedor valido", "", "", "invalido", nil},
		{"sem contato", "Fornecedor valido", "", "", "", nil},
		{"prazo menor que minimo", "Fornecedor valido", "", "11999990000", "", &prazoZero},
		{"prazo maior que maximo", "Fornecedor valido", "", "11999990000", "", &prazoAlto},
	} {
		t.Run(test.nome, func(t *testing.T) {
			if _, err := NovaAtualizacao(test.razao, test.fantasia, test.telefone, test.email, test.prazo); err == nil {
				t.Fatal("atualizacao invalida aceita")
			}
		})
	}
}
