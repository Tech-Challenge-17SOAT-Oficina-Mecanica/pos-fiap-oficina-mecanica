package peca

import "strings"

var (
	ErrAtivoNaoEditavel   = ErroValidacao{"ativo", "ativo nao e alterado por esta operacao; use o DELETE"}
	ErrPrecoObrigatorio   = ErroValidacao{"precoVenda", "precoVenda e obrigatorio e deve ser maior que zero"}
	ErrPrecoCasasDecimais = ErroValidacao{"precoVenda", "precoVenda aceita no maximo 2 casas decimais"}
)

// Atualizacao reune os campos editaveis. Nao inclui codigo, saldo nem ativo: o codigo e
// o identificador de negocio e nao muda, o saldo entra por movimentacao, e a situacao so
// muda pelo DELETE, onde as validacoes de reserva e orcamento existem.
type Atualizacao struct {
	Nome                 string
	Descricao            string
	DescricaoNormalizada string
	CategoriaID          string
	Fabricante           *string
	PrecoVenda           string
	EstoqueMinimo        int64
}

func NovaAtualizacao(nome, descricao, categoriaID string, fabricante, precoVenda *string, estoqueMinimo *int64, ativoInformado bool) (Atualizacao, error) {
	if ativoInformado {
		return Atualizacao{}, ErrAtivoNaoEditavel
	}

	base, err := NovoCadastro(nome, descricao, categoriaID, fabricante, nil, estoqueMinimo)
	if err != nil {
		return Atualizacao{}, err
	}

	if precoVenda == nil || !precoPositivo(*precoVenda) {
		return Atualizacao{}, ErrPrecoObrigatorio
	}
	preco := strings.TrimSpace(*precoVenda)
	if casasDecimais(preco) > 2 {
		return Atualizacao{}, ErrPrecoCasasDecimais
	}

	return Atualizacao{
		Nome:                 base.Nome,
		Descricao:            base.Descricao,
		DescricaoNormalizada: base.DescricaoNormalizada,
		CategoriaID:          base.CategoriaID,
		Fabricante:           base.Fabricante,
		PrecoVenda:           preco,
		EstoqueMinimo:        base.EstoqueMinimo,
	}, nil
}

func precoPositivo(valor string) bool {
	return precoNaoNegativo(valor) && strings.Trim(strings.TrimSpace(valor), "0.") != ""
}

func casasDecimais(valor string) int {
	_, fracao, encontrou := strings.Cut(valor, ".")
	if !encontrou {
		return 0
	}
	return len(fracao)
}
