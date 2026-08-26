package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type registrarProblemaRequest struct {
	Descricao   string `json:"descricao"`
	Observacoes string `json:"observacoes"`
}

type registrarServicosRequest struct {
	Servicos []registrarServicoRequest `json:"servicos"`
}

type registrarServicoRequest struct {
	ServicoID  string `json:"servicoId"`
	Observacao string `json:"observacao"`
}

type servicoRegistradoResponse struct {
	ServicoID     string  `json:"servicoId"`
	Descricao     string  `json:"descricao"`
	ValorUnitario float64 `json:"valorUnitario"`
	Observacao    string  `json:"observacao,omitempty"`
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

func NewRegistrarServicosHandler(useCase application.RegistrarServicos) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ordemServicoID := request.PathValue("osId")
		if !validation.IsUUID(ordemServicoID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		var input registrarServicosRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}
		servicos := make([]domain.ServicoCadastro, 0, len(input.Servicos))
		for index, servico := range input.Servicos {
			if !validation.IsUUID(servico.ServicoID) {
				writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "servicoId invalido", "servicos["+strconv.Itoa(index)+"].servicoId")
				return
			}
			servicos = append(servicos, domain.ServicoCadastro{ServicoID: servico.ServicoID, Observacao: servico.Observacao})
		}
		servicos, err := domain.NovosServicosCadastro(servicos)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "servicos")
			return
		}
		resultado, err := useCase.Execute(request.Context(), ordemServicoID, servicos)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrOrdemServicoNaoEncontrada), errors.Is(err, domain.ErrServicoNaoEncontrado):
				writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "")
			case errors.Is(err, domain.ErrStatusNaoPermiteServico), errors.Is(err, domain.ErrOrcamentoAplicavelNaoEncontrado), errors.Is(err, domain.ErrServicoInativo), errors.Is(err, domain.ErrServicoDuplicado):
				writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
			default:
				writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao registrar servicos", "")
			}
			return
		}
		response := struct {
			OrdemServicoID string `json:"ordemServicoId"`
			Orcamento      struct {
				ID         string  `json:"id"`
				Tipo       string  `json:"tipo"`
				Status     string  `json:"status"`
				ValorTotal float64 `json:"valorTotal"`
			} `json:"orcamento"`
			Servicos []servicoRegistradoResponse `json:"servicos"`
		}{OrdemServicoID: ordemServicoID, Servicos: make([]servicoRegistradoResponse, 0, len(resultado.Servicos))}
		response.Orcamento.ID = resultado.Orcamento.ID
		response.Orcamento.Tipo = resultado.Orcamento.Tipo
		response.Orcamento.Status = resultado.Orcamento.Status
		response.Orcamento.ValorTotal = resultado.Orcamento.ValorTotal
		for _, servico := range resultado.Servicos {
			response.Servicos = append(response.Servicos, servicoRegistradoResponse{ServicoID: servico.ServicoID, Descricao: servico.Descricao, ValorUnitario: servico.ValorUnitario, Observacao: servico.Observacao})
		}
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
