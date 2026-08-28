package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type entregarRequest struct {
	ClienteID   string `json:"clienteId"`
	Observacoes string `json:"observacoes"`
}

type entregarResponse struct {
	OrdemServicoID       string  `json:"ordemServicoId"`
	Status               string  `json:"status"`
	ValorFinal           float64 `json:"valorFinal"`
	ResponsavelEntregaID string  `json:"responsavelEntregaId"`
	ClienteID            string  `json:"clienteId,omitempty"`
	DataEntrega          string  `json:"dataEntrega"`
	Observacoes          string  `json:"observacoes,omitempty"`
}

func NewEntregarHandler(useCase application.Entregar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}

		var input entregarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}

		resultado, err := useCase.Execute(request.Context(), application.EntregarInput{
			OSID: osID, ClienteID: input.ClienteID, Observacoes: input.Observacoes,
			UsuarioID: seguranca.UsuarioID(request.Context()),
		})
		if err != nil {
			writeEntregarError(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(entregarResponse{
			OrdemServicoID: resultado.OrdemServicoID, Status: resultado.Status,
			ValorFinal: resultado.ValorFinal, ResponsavelEntregaID: resultado.ResponsavelEntregaID,
			ClienteID: resultado.ClienteID, DataEntrega: resultado.DataEntrega.Format(time.RFC3339),
			Observacoes: resultado.Observacoes,
		})
	}
}

func writeEntregarError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrClienteIDInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "clienteId")
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada), errors.Is(err, application.ErrClienteNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "")
	case errors.Is(err, domain.ErrOSNaoFinalizada), errors.Is(err, domain.ErrOSJaEntregue), errors.Is(err, domain.ErrValorFinalIndisponivel):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao registrar entrega do veiculo", "")
	}
}
