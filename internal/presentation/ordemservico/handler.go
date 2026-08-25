package ordemservico

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type RegistrarProblemaRelatadoUseCase interface {
	Execute(context.Context, application.RegistrarProblemaRelatadoInput) (domain.OrdemDeServico, error)
}

type registrarProblemaRelatadoRequest struct {
	Descricao   string `json:"descricao"`
	Observacoes string `json:"observacoes"`
}

func NewRegistrarProblemaRelatadoHandler(useCase RegistrarProblemaRelatadoUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
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
		result, err := useCase.Execute(request.Context(), application.RegistrarProblemaRelatadoInput{OrdemServicoID: osID, Descricao: input.Descricao, Observacoes: input.Observacoes})
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"ordemServicoId":        result.ID,
			"problemaRelatado":      map[string]string{"descricao": result.ProblemaRelatado.Descricao, "observacoes": result.ProblemaRelatado.Observacoes},
			"status":                result.Status,
			"dataInicioDiagnostico": result.DataInicioDiagnostico,
		})
	}
}

func writeError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrDescricaoObrigatoria):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "descricao")
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada):
		writeProblem(writer, http.StatusNotFound, "Recurso não encontrado", err.Error(), "")
	case errors.Is(err, application.ErrOrdemServicoForaDeRecebida), errors.Is(err, application.ErrProblemaRelatadoJaRegistrado):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "falha ao registrar problema relatado", "")
	}
}

func writeProblem(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/ordem-servico", Title: title, Status: status, Detail: detail}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}
