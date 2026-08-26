package servico

import (
	"context"
	"errors"
	"strings"

	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

var (
	ErrServicoNaoEncontrado = errors.New("serviço não encontrado")
	ErrPaginaInvalida       = errors.New("pagina deve ser maior ou igual a zero")
	ErrTamanhoInvalido      = errors.New("tamanho deve estar entre 1 e 50")
)

type Filtros struct {
	Nome            string
	IncluirInativos bool
	Pagina          int
	Tamanho         int
}

type Pagina struct {
	Servicos       []domain.Servico
	Pagina         int
	Tamanho        int
	TotalElementos int
	TotalPaginas   int
}

type ConsultarRepository interface {
	Listar(context.Context, Filtros) ([]domain.Servico, int, error)
	BuscarPorID(context.Context, string) (domain.Servico, error)
}

type Consultar struct{ repository ConsultarRepository }

func NewConsultar(repository ConsultarRepository) Consultar { return Consultar{repository: repository} }

func (useCase Consultar) Listar(ctx context.Context, filtros Filtros) (Pagina, error) {
	if filtros.Pagina < 0 {
		return Pagina{}, ErrPaginaInvalida
	}
	if filtros.Tamanho < 1 || filtros.Tamanho > 50 {
		return Pagina{}, ErrTamanhoInvalido
	}
	filtros.Nome = strings.TrimSpace(filtros.Nome)
	servicos, total, err := useCase.repository.Listar(ctx, filtros)
	if err != nil {
		return Pagina{}, err
	}
	if servicos == nil {
		servicos = []domain.Servico{}
	}
	totalPaginas := 0
	if total > 0 {
		totalPaginas = (total + filtros.Tamanho - 1) / filtros.Tamanho
	}
	return Pagina{Servicos: servicos, Pagina: filtros.Pagina, Tamanho: filtros.Tamanho,
		TotalElementos: total, TotalPaginas: totalPaginas}, nil
}

func (useCase Consultar) PorID(ctx context.Context, id string) (domain.Servico, error) {
	return useCase.repository.BuscarPorID(ctx, id)
}
