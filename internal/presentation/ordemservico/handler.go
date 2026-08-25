package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type registrarProblemaRequest struct {
	Descricao   string `json:"descricao"`
	Observacoes string `json:"observacoes"`
}

type registrarProblemaResponse struct {
	ProblemaID  string `json:"problemaId"`
	Descricao   string `json:"descricao"`
	Observacoes string `json:"observacoes,omitempty"`
	Orcamento   struct {
		ID     string `json:"id"`
		Tipo   string `json:"tipo"`
		Status string `json:"status"`
	} `json:"orcamento"`
}

func NewRegistrarProblemaHandler(useCase application.RegistrarProblema) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ordemServicoID := request.PathValue("osId")
		if !validation.IsUUID(ordemServicoID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		var input registrarProblemaRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}
		cadastro, err := domain.NovoProblemaCadastro(input.Descricao, input.Observacoes)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "descricao")
			return
		}
		resultado, err := useCase.Execute(request.Context(), ordemServicoID, cadastro)
		if err != nil {
			if errors.Is(err, application.ErrOrdemServicoNaoEncontrada) {
				writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "osId")
				return
			}
			if errors.Is(err, domain.ErrStatusNaoPermiteProblema) || errors.Is(err, domain.ErrOrcamentoFechado) || errors.Is(err, domain.ErrOrcamentoPrincipalNaoEncontrado) {
				writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao registrar problema", "")
			return
		}
		response := registrarProblemaResponse{ProblemaID: resultado.Problema.ID, Descricao: resultado.Problema.Descricao, Observacoes: resultado.Problema.Observacoes}
		response.Orcamento.ID = resultado.Orcamento.ID
		response.Orcamento.Tipo = resultado.Orcamento.Tipo
		response.Orcamento.Status = resultado.Orcamento.Status
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(response)
	}
}

func writeProblem(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/ordem-servico", Title: title, Status: status, Detail: detail}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}
