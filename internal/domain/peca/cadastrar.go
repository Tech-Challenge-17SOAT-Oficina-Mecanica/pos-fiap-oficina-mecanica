package peca

import (
	"strconv"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const (
	nomeMinimo          = 2
	nomeMaximo          = 150
	descricaoMinima     = 3
	descricaoMaxima     = 120
	fabricanteMaximo    = 150
	UnidadeMedidaPadrao = "UN"
	TipoPeca            = "PECA"
)

// Os erros carregam o parametro que os causou, para a resposta preencher o bloco `erros`
// do problem+json alem do `detail`.
var (
	ErrNomeInvalido          = ErroValidacao{"nome", "nome deve ter entre 2 e 150 caracteres"}
	ErrDescricaoInvalida     = ErroValidacao{"descricao", "descricao deve ter entre 3 e 120 caracteres"}
	ErrCategoriaObrigatoria  = ErroValidacao{"categoriaId", "categoriaId e obrigatorio"}
	ErrFabricanteInvalido    = ErroValidacao{"fabricante", "fabricante deve ter no maximo 150 caracteres"}
	ErrPrecoVendaInvalido    = ErroValidacao{"precoVenda", "precoVenda deve ser maior ou igual a zero"}
	ErrEstoqueMinimoInvalido = ErroValidacao{"estoqueMinimo", "estoqueMinimo deve ser maior ou igual a zero"}
)

// ErroValidacao e comparavel, entao errors.Is continua funcionando por identidade.
type ErroValidacao struct {
	Campo    string
	Mensagem string
}

func (erro ErroValidacao) Error() string { return erro.Mensagem }

// CampoInvalido permite a camada HTTP apontar o parametro culpado na resposta.
func (erro ErroValidacao) CampoInvalido() string { return erro.Campo }

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

func NovoCadastro(nome, descricao, categoriaID string, fabricante, precoVenda *string, estoqueMinimo *int64) (Cadastro, error) {
	cadastro := Cadastro{
		Nome:          strings.TrimSpace(nome),
		Descricao:     validation.ColapsarEspacos(descricao),
		CategoriaID:   strings.TrimSpace(categoriaID),
		UnidadeMedida: UnidadeMedidaPadrao,
	}
	cadastro.DescricaoNormalizada = validation.NormalizarDescricao(cadastro.Descricao)

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
