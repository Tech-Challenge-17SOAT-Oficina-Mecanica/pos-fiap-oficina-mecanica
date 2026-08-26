package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type registrarRequest struct {
	Itens []application.ItemInput `json:"itens"`
}

func NewRegistrarPecasHandler(useCase application.RegistrarItens) http.Handler {
	return newRegistrarHandler(useCase, "PECA")
}

func NewRegistrarInsumosHandler(useCase application.RegistrarItens) http.Handler {
	return newRegistrarHandler(useCase, "INSUMO")
}

func newRegistrarHandler(useCase application.RegistrarItens, tipo string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}

		var input registrarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}

		result, err := useCase.Execute(request.Context(), application.RegistrarInput{
			OSID:      osID,
			Tipo:      tipo,
			Itens:     input.Itens,
			UsuarioID: segurancaUsuarioID(request),
		})
		if err != nil {
			writeUseCaseError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(result)
	})
}

func NewConsultarOrcamentoHandler(useCase application.RegistrarItens) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		results, err := useCase.Consultar(request.Context(), osID)
		if err != nil {
			writeUseCaseError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(results)
	})
}

func segurancaUsuarioID(request *http.Request) string {
	return seguranca.UsuarioID(request.Context())
}

func writeUseCaseError(writer http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	title := "Erro interno"
	switch {
	case errors.Is(err, application.ErrOSNaoEncontrada), errors.Is(err, application.ErrItemNaoEncontrado), errors.Is(err, application.ErrOrcamentoNaoEncontrado):
		status, title = http.StatusNotFound, "Recurso nao encontrado"
	case errors.Is(err, application.ErrItemInativo), errors.Is(err, application.ErrOrcamentoAprovado):
		status, title = http.StatusConflict, "Conflito de estado"
	case errors.Is(err, application.ErrStatusNaoPermiteItens):
		status, title = http.StatusConflict, "Conflito de estado"
	case errors.Is(err, application.ErrItemRepetido):
		status, title = http.StatusBadRequest, "Dados invalidos"
	case strings.Contains(err.Error(), "tipo do item divergente"):
		status, title = http.StatusBadRequest, "Dados invalidos"
	case strings.Contains(err.Error(), "quantidade"), strings.Contains(err.Error(), "itens e obrigatorio"):
		status, title = http.StatusBadRequest, "Dados invalidos"
	}
	writeProblem(writer, status, title, err.Error(), "")
}

func writeProblem(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/ordem-servico", Title: title, Status: status, Detail: detail}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}
