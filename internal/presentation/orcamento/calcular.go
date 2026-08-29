package orcamento

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	orcamentoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type calculoResponse struct {
	OrcamentoID     string  `json:"orcamentoId"`
	OrdemServicoID  string  `json:"ordemServicoId"`
	ValorTotal      float64 `json:"valorTotal"`
	ValorTotalGeral float64 `json:"valorTotalGeral"`
	// Em dias inteiros, sem data exata (RF-ORC-42, RNF-ORC-16).
	EstimativaEntregaDias int `json:"estimativaEntregaDias"`
}

func NewCalcularHandler(useCase orcamento.Calcular) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resultado, err := useCase.Execute(request.Context(), request.PathValue("orcamentoId"))
		if err != nil {
			responderErroCalculo(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(calculoResponse{
			OrcamentoID:     resultado.OrcamentoID,
			OrdemServicoID:  resultado.OrdemServicoID,
			ValorTotal:      resultado.ValorTotal,
			ValorTotalGeral: resultado.ValorTotalGeral,

			EstimativaEntregaDias: resultado.EstimativaDias,
		})
	}
}

func responderErroCalculo(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, orcamento.ErrIdentificadorInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "orcamentoId")
	case errors.Is(err, orcamento.ErrOrcamentoNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Não encontrado", err.Error(), "")
	case errors.Is(err, orcamentoDomain.ErrStatusNaoCalculavel),
		errors.Is(err, orcamentoDomain.ErrComplementarSemPrincipal),
		errors.Is(err, orcamentoDomain.ErrVinculoInvalido):
		writeProblem(writer, http.StatusConflict, "Conflito", err.Error(), "")
	case errors.Is(err, orcamentoDomain.ErrSemItens),
		errors.Is(err, orcamentoDomain.ErrItemInvalido):
		writeProblem(writer, http.StatusConflict, "Conflito", err.Error(), "itens")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "falha ao calcular o orçamento", "")
	}
}
