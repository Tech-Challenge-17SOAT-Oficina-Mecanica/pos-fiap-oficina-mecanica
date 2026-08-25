package peca

import (
	"errors"
	"strconv"
	"strings"
)

const (
	nomeMinimo          = 2
	nomeMaximo          = 150
	descricaoMinima     = 3
	descricaoMaxima     = 500
	fabricanteMaximo    = 150
	UnidadeMedidaPadrao = "UN"
	TipoPeca            = "PECA"
)

var removedorAcentos = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
	"Á", "A", "À", "A", "Â", "A", "Ã", "A", "Ä", "A",
	"É", "E", "È", "E", "Ê", "E", "Ë", "E",
	"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
	"Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O", "Ö", "O",
	"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
	"Ç", "C", "Ñ", "N",
)

var (
	ErrNomeInvalido          = errors.New("nome deve ter entre 2 e 150 caracteres")
	ErrDescricaoInvalida     = errors.New("descricao deve ter entre 3 e 500 caracteres")
	ErrCategoriaObrigatoria  = errors.New("categoriaId e obrigatorio")
	ErrFabricanteInvalido    = errors.New("fabricante deve ter no maximo 150 caracteres")
	ErrPrecoVendaInvalido    = errors.New("precoVenda deve ser maior ou igual a zero")
	ErrEstoqueMinimoInvalido = errors.New("estoqueMinimo deve ser maior ou igual a zero")
)

// Cadastro reune os dados que o cliente informa. O codigo, o id, o saldo e a data de
// criacao nao entram aqui: sao gerados pelo sistema ou por outros fluxos.
type Cadastro struct {
	Nome                 string
	Descricao            string
	DescricaoNormalizada string
	CategoriaID          string
	Fabricante           *string
	PrecoVenda           *string
	EstoqueMinimo        int64
	UnidadeMedida        string
}

// NormalizarDescricao devolve a forma usada na regra de duplicidade: sem acento, sem
// espaco duplo e em minusculas.
func NormalizarDescricao(valor string) string {
	semAcento := removedorAcentos.Replace(valor)
	return strings.ToLower(strings.Join(strings.Fields(semAcento), " "))
}

func NovoCadastro(nome, descricao, categoriaID string, fabricante, precoVenda *string, estoqueMinimo *int64) (Cadastro, error) {
	cadastro := Cadastro{
		Nome:          strings.TrimSpace(nome),
		Descricao:     strings.Join(strings.Fields(descricao), " "),
		CategoriaID:   strings.TrimSpace(categoriaID),
		UnidadeMedida: UnidadeMedidaPadrao,
	}
	cadastro.DescricaoNormalizada = NormalizarDescricao(cadastro.Descricao)

	if tamanho := len([]rune(cadastro.Nome)); tamanho < nomeMinimo || tamanho > nomeMaximo {
		return Cadastro{}, ErrNomeInvalido
	}
	if tamanho := len([]rune(cadastro.Descricao)); tamanho < descricaoMinima || tamanho > descricaoMaxima {
		return Cadastro{}, ErrDescricaoInvalida
	}
	if cadastro.CategoriaID == "" {
		return Cadastro{}, ErrCategoriaObrigatoria
	}

	if fabricante != nil {
		limpo := strings.TrimSpace(*fabricante)
		if len([]rune(limpo)) > fabricanteMaximo {
			return Cadastro{}, ErrFabricanteInvalido
		}
		if limpo != "" {
			cadastro.Fabricante = &limpo
		}
	}
	if precoVenda != nil {
		if !precoNaoNegativo(*precoVenda) {
			return Cadastro{}, ErrPrecoVendaInvalido
		}
		cadastro.PrecoVenda = precoVenda
	}
	if estoqueMinimo != nil {
		if *estoqueMinimo < 0 {
			return Cadastro{}, ErrEstoqueMinimoInvalido
		}
		cadastro.EstoqueMinimo = *estoqueMinimo
	}

	return cadastro, nil
}

func precoNaoNegativo(valor string) bool {
	preco, err := strconv.ParseFloat(strings.TrimSpace(valor), 64)
	return err == nil && preco >= 0
}
