package peca

import (
	"errors"
	"strings"
	"testing"
)

const categoriaTeste = "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94"

func ptr(valor string) *string { return &valor }

func TestNovaAtualizacaoValida(t *testing.T) {
	estoque := int64(6)
	atualizacao, err := NovaAtualizacao("Pastilha de freio", "Pastilha  de  freio  cerâmica",
		categoriaTeste, ptr("Bosch"), ptr("199.90"), &estoque, false)
	if err != nil {
		t.Fatal(err)
	}

	if atualizacao.Descricao != "Pastilha de freio cerâmica" {
		t.Fatalf("descricao = %q", atualizacao.Descricao)
	}
	if atualizacao.DescricaoNormalizada != "pastilha de freio ceramica" {
		t.Fatalf("descricaoNormalizada = %q", atualizacao.DescricaoNormalizada)
	}
	if atualizacao.PrecoVenda != "199.90" {
		t.Fatalf("precoVenda = %q", atualizacao.PrecoVenda)
	}
	if atualizacao.EstoqueMinimo != 6 {
		t.Fatalf("estoqueMinimo = %d", atualizacao.EstoqueMinimo)
	}
}

// A situação só muda pelo DELETE, onde as validações de saldo e orçamento existem.
// Vale para qualquer valor: o que importa é o campo ter vindo no corpo.
func TestNovaAtualizacaoRecusaAtivo(t *testing.T) {
	_, err := NovaAtualizacao("Peca", "Descricao valida", categoriaTeste, nil, ptr("10"), nil, true)
	if !errors.Is(err, ErrAtivoNaoEditavel) {
		t.Fatalf("informar ativo deveria ser recusado, veio %v", err)
	}

	if _, err := NovaAtualizacao("Peca", "Descricao valida", categoriaTeste, nil, ptr("10"), nil, false); err != nil {
		t.Fatalf("omitir ativo deveria passar, veio %v", err)
	}
}

func TestNovaAtualizacaoPreco(t *testing.T) {
	casos := []struct {
		nome     string
		preco    *string
		esperado error
	}{
		{"ausente", nil, ErrPrecoObrigatorio},
		{"zero", ptr("0"), ErrPrecoObrigatorio},
		{"zero com casas", ptr("0.00"), ErrPrecoObrigatorio},
		{"negativo", ptr("-5"), ErrPrecoObrigatorio},
		{"nao numerico", ptr("abc"), ErrPrecoObrigatorio},
		{"tres casas", ptr("10.123"), ErrPrecoCasasDecimais},
		{"duas casas", ptr("10.12"), nil},
		{"uma casa", ptr("10.1"), nil},
		{"inteiro", ptr("10"), nil},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			_, err := NovaAtualizacao("Peca", "Descricao valida", categoriaTeste, nil, caso.preco, nil, false)
			if !errors.Is(err, caso.esperado) {
				t.Fatalf("erro = %v, esperado %v", err, caso.esperado)
			}
		})
	}
}

// O limite de descrição precisa ser o mesmo do cadastro, senão dá para cadastrar uma
// peça que depois não pode ser atualizada sem encurtar o texto.
func TestLimiteDeDescricaoIgualAoCadastro(t *testing.T) {
	noLimite := strings.Repeat("a", 120)
	acima := strings.Repeat("a", 121)

	if _, err := NovaAtualizacao("Peca", noLimite, categoriaTeste, nil, ptr("10"), nil, false); err != nil {
		t.Fatalf("120 caracteres deveriam passar na atualização: %v", err)
	}
	if _, err := NovoCadastro("Peca", noLimite, categoriaTeste, nil, nil, nil); err != nil {
		t.Fatalf("120 caracteres deveriam passar no cadastro: %v", err)
	}
	if _, err := NovaAtualizacao("Peca", acima, categoriaTeste, nil, ptr("10"), nil, false); !errors.Is(err, ErrDescricaoInvalida) {
		t.Fatalf("121 caracteres deveriam falhar na atualização, veio %v", err)
	}
	if _, err := NovoCadastro("Peca", acima, categoriaTeste, nil, nil, nil); !errors.Is(err, ErrDescricaoInvalida) {
		t.Fatalf("121 caracteres deveriam falhar no cadastro, veio %v", err)
	}
}
