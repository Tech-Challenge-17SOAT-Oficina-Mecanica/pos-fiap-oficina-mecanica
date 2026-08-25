package fornecedor

import "testing"

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
		{"Fornecedor válido", string(make([]byte, 121)), "CNPJ", nil},
		{"Fornecedor válido", "", "RG", nil},
		{"Fornecedor válido", "", "CNPJ", &prazoInvalido},
		{"Fornecedor válido", "", "CPF", nil},
	} {
		if _, err := NovoCadastro(test.razao, test.fantasia, "04.252.011/0001-10", test.tipo, "11999990000", "", test.prazo); err == nil {
			t.Fatal("cadastro inválido aceito")
		}
	}
}
