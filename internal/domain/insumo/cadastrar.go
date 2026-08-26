package insumo

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const (
	nomeMinimo      = 2
	nomeMaximo      = 150
	descricaoMinima = 3
	descricaoMaxima = 500
)

// UnidadesMedida e fechado de proposito: nao ha conversao entre unidades, cada uma e um
// item independente no catalogo.
var UnidadesMedida = []string{"UN", "L", "ML", "KG", "G", "M"}

var (
	ErrNomeInvalido          = errors.New("nome deve ter entre 2 e 150 caracteres")
	ErrDescricaoInvalida     = errors.New("descricao deve ter entre 3 e 500 caracteres")
	ErrCategoriaObrigatoria  = errors.New("categoriaId e obrigatorio")
	ErrUnidadeInvalida       = errors.New("unidadeMedida deve ser uma de: UN, L, ML, KG, G, M")
	ErrCustoInvalido         = errors.New("custoUnitario e obrigatorio e nao pode ser negativo")
	ErrEstoqueMinimoInvalido = errors.New("estoqueMinimo nao pode ser negativo")
)

type Cadastro struct {
	Nome                 string
	Descricao            string
	DescricaoNormalizada string
	CategoriaID          string
	UnidadeMedida        string
	CustoUnitario        string
	EstoqueMinimo        string
}

func NovoCadastro(nome, descricao, categoriaID, unidadeMedida string, custoUnitario, estoqueMinimo *string) (Cadastro, error) {
	cadastro := Cadastro{
		Nome:          strings.TrimSpace(nome),
		Descricao:     validation.ColapsarEspacos(descricao),
		CategoriaID:   strings.TrimSpace(categoriaID),
		UnidadeMedida: strings.ToUpper(strings.TrimSpace(unidadeMedida)),
		EstoqueMinimo: "0",
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
	if !slices.Contains(UnidadesMedida, cadastro.UnidadeMedida) {
		return Cadastro{}, ErrUnidadeInvalida
	}

	if custoUnitario == nil || !decimalNaoNegativo(*custoUnitario) {
		return Cadastro{}, ErrCustoInvalido
	}
	cadastro.CustoUnitario = strings.TrimSpace(*custoUnitario)

	if estoqueMinimo != nil {
		if !decimalNaoNegativo(*estoqueMinimo) {
			return Cadastro{}, ErrEstoqueMinimoInvalido
		}
		cadastro.EstoqueMinimo = strings.TrimSpace(*estoqueMinimo)
	}

	return cadastro, nil
}

func decimalNaoNegativo(valor string) bool {
	numero, err := strconv.ParseFloat(strings.TrimSpace(valor), 64)
	return err == nil && numero >= 0
}
