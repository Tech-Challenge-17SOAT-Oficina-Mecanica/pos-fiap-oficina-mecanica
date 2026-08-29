package insumo

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
)

type processamentoRequest struct {
	OrdemServicoID string                     `json:"ordemServicoId"`
	FornecedorID   string                     `json:"fornecedorId"`
	Itens          []insumo.ItemProcessamento `json:"itens"`
}

func NewSolicitarCompraEReservarInsumosHandler(useCase insumo.SolicitarCompraEReservarInsumos) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		corpo, err := io.ReadAll(request.Body)
		if err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}

		var payload processamentoRequest
		decoder := json.NewDecoder(bytes.NewReader(corpo))
		decoder.UseNumber()
		if err := decoder.Decode(&payload); err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}

		hash := sha256.Sum256(corpo)
		resultado, err := useCase.Execute(request.Context(), insumo.SolicitacaoCompraReserva{
			IdempotencyKey: request.Header.Get("Idempotency-Key"),
			HashRequisicao: hex.EncodeToString(hash[:]),
			OrdemServicoID: payload.OrdemServicoID,
			FornecedorID:   payload.FornecedorID,
			Itens:          payload.Itens,
		})
		if err != nil {
			responderErroProcessamento(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if resultado.Reprocessado {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusCreated)
		}
		_ = json.NewEncoder(writer).Encode(resultado)
	}
}

func responderErroProcessamento(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, insumo.ErrIdempotencyKeyObrigatoria):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "Idempotency-Key")
	case errors.Is(err, insumo.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "ordemServicoId")
	case errors.Is(err, insumo.ErrFornecedorIdentificador):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "fornecedorId")
	case errors.Is(err, insumo.ErrFornecedorNaoEncontrado):
		problema(writer, http.StatusNotFound, "Não encontrado", err.Error(), "fornecedorId")
	case errors.Is(err, insumo.ErrFornecedorInativo):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "fornecedorId")
	case errors.Is(err, insumo.ErrItemObrigatorio):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "itens")
	case errors.Is(err, insumo.ErrItemRepetido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "itens")
	case errors.Is(err, insumo.ErrQuantidadeProcessamento):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "quantidade")
	case errors.Is(err, insumo.ErrItemIdentificador):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "itemId")
	case errors.Is(err, insumo.ErrInsumoNaoEncontrado):
		problema(writer, http.StatusNotFound, "Não encontrado", err.Error(), "itemId")
	case errors.Is(err, insumo.ErrItemProcessamentoInvalido):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "itemId")
	case errors.Is(err, insumo.ErrOrdemServicoNaoEncontrada):
		problema(writer, http.StatusNotFound, "Não encontrado", err.Error(), "ordemServicoId")
	case errors.Is(err, insumo.ErrOrdemServicoInvalida):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "ordemServicoId")
	case errors.Is(err, insumo.ErrProcessamentoDuplicado):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "itens")
	case errors.Is(err, insumo.ErrIdempotencyKeyEmUso):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "Idempotency-Key")
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao processar insumos", "")
	}
}
