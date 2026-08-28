package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type registrarProblemaRelatadoRequest struct {
	Descricao   string `json:"descricao"`
	Observacoes string `json:"observacoes"`
}

type problemaRelatadoResponse struct {
	Descricao   string `json:"descricao"`
	Observacoes string `json:"observacoes,omitempty"`
}

func NewRegistrarProblemaRelatadoHandler(useCase application.RegistrarProblemaRelatado) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ordemServicoID := request.PathValue("osId")
		if !validation.IsUUID(ordemServicoID) {
			writeProblem(writer, http.StatusBadRequest, "Dados inválidos", "osId inválido", "osId")
			return
		}
		var input registrarProblemaRelatadoRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		resultado, err := useCase.Execute(request.Context(), application.RegistrarProblemaRelatadoInput{
			OrdemServicoID: ordemServicoID, Descricao: input.Descricao, Observacoes: input.Observacoes,
		})
		if err != nil {
			responderErroProblemaRelatado(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(struct {
			OrdemServicoID        string                   `json:"ordemServicoId"`
			ProblemaRelatado      problemaRelatadoResponse `json:"problemaRelatado"`
			Status                string                   `json:"status"`
			DataInicioDiagnostico *time.Time               `json:"dataInicioDiagnostico"`
		}{resultado.ID, problemaRelatadoResponse{Descricao: resultado.ProblemaRelatado.Descricao, Observacoes: resultado.ProblemaRelatado.Observacoes}, resultado.Status, resultado.DataInicioDiagnostico})
	}
}

func responderErroProblemaRelatado(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrDescricaoProblemaRelatadoObrigatoria):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "descricao")
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada):
		writeProblem(writer, http.StatusNotFound, "Recurso não encontrado", err.Error(), "osId")
	case errors.Is(err, application.ErrOrdemServicoForaDeRecebida), errors.Is(err, application.ErrProblemaRelatadoJaRegistrado):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "falha ao registrar problema relatado", "")
	}
}
