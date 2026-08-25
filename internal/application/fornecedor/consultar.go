package fornecedor

import (
	"context"
	"errors"
	"strings"
	"unicode"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

const (
	TamanhoPaginaPadrao = 20
	TamanhoPaginaMaximo = 50
)

var (
	ErrFornecedorNaoEncontrado = errors.New("fornecedor nao encontrado")
	ErrConsultaInvalida        = errors.New("parametros de consulta invalidos")
)

type FiltrosConsulta struct {
	Nome            string
	Documento       string
	IncluirInativos bool
	Pagina          int
	Tamanho         int
}

type ResultadoConsulta struct {
	Data           []domain.Fornecedor
	Pagina         int
	Tamanho        int
	TotalElementos int
	TotalPaginas   int
}

type ConsultaRepository interface {
	Listar(context.Context, FiltrosConsulta) ([]domain.Fornecedor, int, error)
	BuscarPorID(context.Context, string) (domain.Fornecedor, error)
}

type ConsultarFornecedores struct {
	repository ConsultaRepository
}

func NewConsultarFornecedores(repository ConsultaRepository) ConsultarFornecedores {
	return ConsultarFornecedores{repository: repository}
}

func (useCase ConsultarFornecedores) Execute(ctx context.Context, filtros FiltrosConsulta) (ResultadoConsulta, error) {
	filtros.Normalizar()
	if !filtros.Validos() {
		return ResultadoConsulta{}, ErrConsultaInvalida
	}
	fornecedores, total, err := useCase.repository.Listar(ctx, filtros)
	if err != nil {
		return ResultadoConsulta{}, err
	}
	return ResultadoConsulta{
		Data:           fornecedores,
		Pagina:         filtros.Pagina,
		Tamanho:        filtros.Tamanho,
		TotalElementos: total,
		TotalPaginas:   totalPaginas(total, filtros.Tamanho),
	}, nil
}

type ConsultarFornecedorPorID struct {
	repository ConsultaRepository
}

func NewConsultarFornecedorPorID(repository ConsultaRepository) ConsultarFornecedorPorID {
	return ConsultarFornecedorPorID{repository: repository}
}

func (useCase ConsultarFornecedorPorID) Execute(ctx context.Context, fornecedorID string) (domain.Fornecedor, error) {
	fornecedorID = strings.TrimSpace(fornecedorID)
	if fornecedorID == "" {
		return domain.Fornecedor{}, ErrConsultaInvalida
	}
	fornecedor, err := useCase.repository.BuscarPorID(ctx, fornecedorID)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func (filtros *FiltrosConsulta) Normalizar() {
	filtros.Nome = strings.TrimSpace(filtros.Nome)
	filtros.Documento = strings.TrimSpace(filtros.Documento)
	if filtros.Tamanho == 0 {
		filtros.Tamanho = TamanhoPaginaPadrao
	}
}

func (filtros FiltrosConsulta) Validos() bool {
	return filtros.Pagina >= 0 && filtros.Tamanho > 0 && filtros.Tamanho <= TamanhoPaginaMaximo && documentoConsultaValido(filtros.Documento)
}

func documentoConsultaValido(documento string) bool {
	if documento == "" {
		return true
	}
	for _, character := range documento {
		if !unicode.IsDigit(character) {
			return false
		}
	}
	return true
}

func totalPaginas(total, tamanho int) int {
	if total == 0 {
		return 0
	}
	return (total + tamanho - 1) / tamanho
}
