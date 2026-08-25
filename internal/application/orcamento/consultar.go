package orcamento

import (
	"context"
	"errors"
	"strings"

	clienteDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

var (
	ErrCriterioObrigatorio    = errors.New("informe orcamentoId ou documento")
	ErrOrcamentoNaoEncontrado = errors.New("orçamento não encontrado")
	ErrClienteNaoEncontrado   = errors.New("cliente não encontrado")
	ErrPaginacaoInvalida      = errors.New("pagina deve ser maior ou igual a zero e tamanho deve estar entre 1 e 100")
)

type Repository interface {
	BuscarPorID(context.Context, string) (domain.Consulta, error)
	BuscarPorDocumento(context.Context, string, int, int) ([]domain.Consulta, int, error)
}

type ConsultarInput struct {
	OrcamentoID string
	Documento   string
	Pagina      int
	Tamanho     int
}

type Resultado struct {
	Consulta       *domain.Consulta
	Data           []domain.Consulta
	Pagina         int
	Tamanho        int
	TotalElementos int
	TotalPaginas   int
}

type Consultar struct{ repository Repository }

func NewConsultar(repository Repository) Consultar { return Consultar{repository: repository} }

func (useCase Consultar) Execute(ctx context.Context, input ConsultarInput) (Resultado, error) {
	input.OrcamentoID = strings.TrimSpace(input.OrcamentoID)
	input.Documento = strings.TrimSpace(input.Documento)
	if input.OrcamentoID == "" && input.Documento == "" {
		return Resultado{}, ErrCriterioObrigatorio
	}
	if input.Pagina < 0 || input.Tamanho < 1 || input.Tamanho > 100 {
		return Resultado{}, ErrPaginacaoInvalida
	}
	if input.OrcamentoID != "" {
		consulta, err := useCase.repository.BuscarPorID(ctx, input.OrcamentoID)
		if err != nil {
			return Resultado{}, err
		}
		return Resultado{Consulta: &consulta}, nil
	}
	documento, err := clienteDomain.DocumentoParaConsulta(input.Documento)
	if err != nil {
		return Resultado{}, err
	}
	data, total, err := useCase.repository.BuscarPorDocumento(ctx, documento, input.Pagina*input.Tamanho, input.Tamanho)
	if err != nil {
		return Resultado{}, err
	}
	totalPaginas := 0
	if total > 0 {
		totalPaginas = (total + input.Tamanho - 1) / input.Tamanho
	}
	return Resultado{Data: data, Pagina: input.Pagina, Tamanho: input.Tamanho, TotalElementos: total, TotalPaginas: totalPaginas}, nil
}
