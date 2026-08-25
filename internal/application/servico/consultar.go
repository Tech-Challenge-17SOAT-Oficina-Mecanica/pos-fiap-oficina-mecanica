package servico

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

const (
	TamanhoPaginaPadrao = 20
	TamanhoPaginaMaximo = 50
)

var (
	ErrServicoNaoEncontrado = errors.New("serviço não encontrado")
	ErrConsultaInvalida     = errors.New("parâmetros de consulta inválidos")
)

type FiltrosConsulta struct {
	Nome            string
	IncluirInativos bool
	Pagina          int
	Tamanho         int
}

type ResultadoConsulta struct {
	Data           []domain.Servico
	Pagina         int
	Tamanho        int
	TotalElementos int
	TotalPaginas   int
}

type ConsultaRepository interface {
	Listar(context.Context, FiltrosConsulta) ([]domain.Servico, int, error)
	BuscarPorID(context.Context, string) (domain.Servico, error)
}

type ConsultarServicos struct {
	repository ConsultaRepository
}

func NewConsultarServicos(repository ConsultaRepository) ConsultarServicos {
	return ConsultarServicos{repository: repository}
}

func (uc ConsultarServicos) Execute(ctx context.Context, filtros FiltrosConsulta) (ResultadoConsulta, error) {
	filtros.Normalizar()
	if !filtros.Validos() {
		return ResultadoConsulta{}, ErrConsultaInvalida
	}
	servicos, total, err := uc.repository.Listar(ctx, filtros)
	if err != nil {
		return ResultadoConsulta{}, err
	}
	return ResultadoConsulta{
		Data:           servicos,
		Pagina:         filtros.Pagina,
		Tamanho:        filtros.Tamanho,
		TotalElementos: total,
		TotalPaginas:   totalPaginas(total, filtros.Tamanho),
	}, nil
}

type ConsultarServicoPorID struct {
	repository ConsultaRepository
}

func NewConsultarServicoPorID(repository ConsultaRepository) ConsultarServicoPorID {
	return ConsultarServicoPorID{repository: repository}
}

func (uc ConsultarServicoPorID) Execute(ctx context.Context, servicoID string) (domain.Servico, error) {
	servicoID = strings.TrimSpace(servicoID)
	if servicoID == "" {
		return domain.Servico{}, ErrConsultaInvalida
	}
	return uc.repository.BuscarPorID(ctx, servicoID)
}

func (f *FiltrosConsulta) Normalizar() {
	f.Nome = strings.TrimSpace(f.Nome)
	if f.Tamanho == 0 {
		f.Tamanho = TamanhoPaginaPadrao
	}
}

func (f FiltrosConsulta) Validos() bool {
	return f.Pagina >= 0 && f.Tamanho > 0 && f.Tamanho <= TamanhoPaginaMaximo
}

func totalPaginas(total, tamanho int) int {
	if total == 0 {
		return 0
	}
	return (total + tamanho - 1) / tamanho
}
