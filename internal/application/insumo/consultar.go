package insumo

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const descricaoMinima = 2

var decimalPositivo = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]{1,3})?$`)

var (
	ErrFiltroObrigatorio     = errors.New("informe codigo, descricao ou categoriaId")
	ErrDescricaoCurta        = errors.New("descricao deve ter no minimo 2 caracteres")
	ErrQuantidadeInvalida    = errors.New("quantidadeDesejada deve ser maior que zero e ter ate 3 casas decimais")
	ErrQuantidadeObrigatoria = errors.New("quantidadeDesejada e obrigatoria quando somenteDisponiveis for true")
	ErrInsumoNaoEncontrado   = errors.New("insumo nao encontrado")
)

type FiltrosConsulta struct {
	Codigo             string
	Descricao          string
	CategoriaID        string
	QuantidadeDesejada *string
	SomenteDisponiveis bool
	IncluirInativos    bool
}

type ConsultarRepository interface {
	BuscarPorFiltro(ctx context.Context, filtros FiltrosConsulta, limite, deslocamento int) ([]insumo.Insumo, int, error)
	BuscarPorID(ctx context.Context, id string) (insumo.Insumo, error)
}

type ResultadoConsulta struct {
	Insumos        []insumo.Insumo
	TotalElementos int
}

type ConsultarInsumos struct {
	repository ConsultarRepository
}

func NewConsultarInsumos(repository ConsultarRepository) ConsultarInsumos {
	return ConsultarInsumos{repository: repository}
}

func (useCase ConsultarInsumos) Execute(ctx context.Context, filtros FiltrosConsulta, limite, deslocamento int) (ResultadoConsulta, error) {
	filtros, err := normalizarConsulta(filtros)
	if err != nil {
		return ResultadoConsulta{}, err
	}
	itens, total, err := useCase.repository.BuscarPorFiltro(ctx, filtros, limite, deslocamento)
	if err != nil {
		return ResultadoConsulta{}, err
	}
	return ResultadoConsulta{Insumos: itens, TotalElementos: total}, nil
}

func (useCase ConsultarInsumos) BuscarPorID(ctx context.Context, id string) (insumo.Insumo, error) {
	id = strings.TrimSpace(id)
	if !validation.IsUUID(id) {
		return insumo.Insumo{}, ErrIdentificadorInvalido
	}
	return useCase.repository.BuscarPorID(ctx, id)
}

func normalizarConsulta(filtros FiltrosConsulta) (FiltrosConsulta, error) {
	filtros.Codigo = strings.TrimSpace(filtros.Codigo)
	filtros.Descricao = strings.TrimSpace(filtros.Descricao)
	filtros.CategoriaID = strings.TrimSpace(filtros.CategoriaID)
	if filtros.QuantidadeDesejada != nil {
		quantidade := strings.TrimSpace(*filtros.QuantidadeDesejada)
		filtros.QuantidadeDesejada = &quantidade
	}

	if filtros.Codigo == "" && filtros.Descricao == "" && filtros.CategoriaID == "" {
		return FiltrosConsulta{}, ErrFiltroObrigatorio
	}
	if filtros.Descricao != "" && len([]rune(filtros.Descricao)) < descricaoMinima {
		return FiltrosConsulta{}, ErrDescricaoCurta
	}
	if filtros.CategoriaID != "" && !validation.IsUUID(filtros.CategoriaID) {
		return FiltrosConsulta{}, ErrIdentificadorInvalido
	}
	if filtros.QuantidadeDesejada != nil && !QuantidadeValida(*filtros.QuantidadeDesejada) {
		return FiltrosConsulta{}, ErrQuantidadeInvalida
	}
	if filtros.SomenteDisponiveis && filtros.QuantidadeDesejada == nil {
		return FiltrosConsulta{}, ErrQuantidadeObrigatoria
	}
	return filtros, nil
}

func QuantidadeValida(valor string) bool {
	if !decimalPositivo.MatchString(valor) {
		return false
	}
	return valor != "0" && valor != "0.0" && valor != "0.00" && valor != "0.000"
}
