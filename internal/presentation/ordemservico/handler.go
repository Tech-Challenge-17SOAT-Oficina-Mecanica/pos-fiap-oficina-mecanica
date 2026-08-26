package ordemservico

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type CriarUseCase interface {
	Execute(context.Context, application.CriarInput) (domain.OrdemDeServico, error)
}

type criarRequest struct {
	ClienteID string `json:"clienteId"`
	VeiculoID string `json:"veiculoId"`
}

type criarResponse struct {
	OrdemServicoID string `json:"ordemServicoId"`
	ClienteID      string `json:"clienteId"`
	VeiculoID      string `json:"veiculoId"`
	Status         string `json:"status"`
	CriadaEm       string `json:"criadaEm"`
}

func NewCriarHandler(useCase CriarUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input criarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido")
			return
		}
		ordem, err := useCase.Execute(request.Context(), application.CriarInput{ClienteID: input.ClienteID, VeiculoID: input.VeiculoID})
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(criarResponse{
			OrdemServicoID: ordem.ID,
			ClienteID:      ordem.ClienteID,
			VeiculoID:      ordem.VeiculoID,
			Status:         ordem.Status,
			CriadaEm:       ordem.CriadaEm.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func writeError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrClienteIDObrigatorio), errors.Is(err, application.ErrVeiculoIDObrigatorio),
		errors.Is(err, application.ErrClienteIDInvalido), errors.Is(err, application.ErrVeiculoIDInvalido):
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
	case errors.Is(err, application.ErrClienteNaoEncontrado), errors.Is(err, application.ErrVeiculoNaoEncontrado):
		problem(writer, http.StatusNotFound, "Não encontrado", err.Error())
	case errors.Is(err, application.ErrVeiculoNaoVinculadoCliente):
		problem(writer, http.StatusConflict, "Conflito", err.Error())
	default:
		problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao criar Ordem de Serviço")
	}
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/ordens-servico", Title: title, Status: status, Detail: detail})
}
