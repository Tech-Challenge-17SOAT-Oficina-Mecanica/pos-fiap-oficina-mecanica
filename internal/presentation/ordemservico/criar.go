package ordemservico

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type CriarUseCase interface {
	Execute(context.Context, application.CriarInput) (domain.OrdemDeServico, error)
}

type criarRequest struct {
	ClienteID string `json:"clienteId"`
	VeiculoID string `json:"veiculoId"`
}

func NewCriarHandler(useCase CriarUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input criarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		ordem, err := useCase.Execute(request.Context(), application.CriarInput{ClienteID: input.ClienteID, VeiculoID: input.VeiculoID})
		if err != nil {
			writeCriarError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(struct {
			OrdemServicoID string `json:"ordemServicoId"`
			ClienteID      string `json:"clienteId"`
			VeiculoID      string `json:"veiculoId"`
			Status         string `json:"status"`
			CriadaEm       string `json:"criadaEm"`
		}{ordem.ID, ordem.ClienteID, ordem.VeiculoID, ordem.Status, ordem.CriadaEm.Format("2006-01-02T15:04:05Z07:00")})
	}
}

func writeCriarError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrClienteIDObrigatorio), errors.Is(err, application.ErrClienteIDInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "clienteId")
	case errors.Is(err, application.ErrVeiculoIDObrigatorio), errors.Is(err, application.ErrVeiculoIDInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "veiculoId")
	case errors.Is(err, application.ErrClienteNaoEncontrado), errors.Is(err, application.ErrVeiculoNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso não encontrado", err.Error(), "")
	case errors.Is(err, application.ErrVeiculoNaoVinculadoCliente):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "veiculoId")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "falha ao criar ordem de serviço", "")
	}
}
