package peca

import (
	"context"
	"errors"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const descricaoMinima = 2

var (
	ErrFiltroObrigatorio     = errors.New("informe codigo ou descricao")
	ErrDescricaoCurta        = errors.New("descricao deve ter no minimo 2 caracteres")
	ErrQuantidadeInvalida    = errors.New("quantidadeDesejada deve ser maior que zero")
	ErrIdentificadorInvalido = errors.New("identificador deve ser um UUID valido")
	ErrNaoEncontrada         = errors.New("nenhuma peca corresponde aos parametros informados")
)

type Filtros struct {
	Codigo             string
	Descricao          string
	CategoriaID        string
	Fabricante         string
	SomenteDisponiveis bool
	IncluirInativos    bool
	QuantidadeDesejada *int64
}

type Repository interface {
	BuscarPorFiltro(ctx context.Context, filtros Filtros, limite, deslocamento int) ([]peca.Peca, int, error)
	BuscarPorID(ctx context.Context, id string) (peca.Peca, error)
}

type Resultado struct {
	Pecas          []peca.Peca
	TotalElementos int
}

type ConsultarPecas struct {
	repository Repository
}

func NewConsultarPecas(repository Repository) ConsultarPecas {
	return ConsultarPecas{repository: repository}
}

func (useCase ConsultarPecas) Execute(ctx context.Context, filtros Filtros, limite, deslocamento int) (Resultado, error) {
	filtros, err := normalizar(filtros)
	if err != nil {
		return Resultado{}, err
	}

	pecas, total, err := useCase.repository.BuscarPorFiltro(ctx, filtros, limite, deslocamento)
	if err != nil {
		return Resultado{}, err
	}
	return Resultado{Pecas: pecas, TotalElementos: total}, nil
}

func (useCase ConsultarPecas) BuscarPorID(ctx context.Context, id string) (peca.Peca, error) {
	if !validation.IsUUID(strings.TrimSpace(id)) {
		return peca.Peca{}, ErrIdentificadorInvalido
	}
	return useCase.repository.BuscarPorID(ctx, strings.TrimSpace(id))
}

func normalizar(filtros Filtros) (Filtros, error) {
	filtros.Codigo = strings.TrimSpace(filtros.Codigo)
	filtros.Descricao = strings.TrimSpace(filtros.Descricao)
	filtros.Fabricante = strings.TrimSpace(filtros.Fabricante)
	filtros.CategoriaID = strings.TrimSpace(filtros.CategoriaID)

	if filtros.Codigo == "" && filtros.Descricao == "" {
		return Filtros{}, ErrFiltroObrigatorio
	}
	if filtros.Descricao != "" && len([]rune(filtros.Descricao)) < descricaoMinima {
		return Filtros{}, ErrDescricaoCurta
	}
	if filtros.CategoriaID != "" && !validation.IsUUID(filtros.CategoriaID) {
		return Filtros{}, ErrIdentificadorInvalido
	}
	if filtros.QuantidadeDesejada != nil && *filtros.QuantidadeDesejada <= 0 {
		return Filtros{}, ErrQuantidadeInvalida
	}
	return filtros, nil
}
