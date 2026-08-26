package servico

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type CadastrarUseCase interface {
	Execute(context.Context, domain.NovoServicoInput, string) (domain.Servico, error)
}

type cadastrarRequest struct {
	Nome                 string       `json:"nome"`
	Descricao            string       `json:"descricao"`
	Valor                *json.Number `json:"valor"`
	TempoEstimadoMinutos int          `json:"tempoEstimadoMinutos"`
}

type servicoResponse struct {
	ID                   string      `json:"id"`
	Codigo               string      `json:"codigo"`
	Nome                 string      `json:"nome"`
	Descricao            string      `json:"descricao,omitempty"`
	Valor                json.Number `json:"valor"`
	TempoEstimadoMinutos int         `json:"tempoEstimadoMinutos"`
	Ativo                bool        `json:"ativo"`
	Version              int         `json:"version"`
	DataCriacao          string      `json:"dataCriacao"`
}

func NewCadastrarHandler(useCase CadastrarUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input cadastrarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		if input.Valor == nil {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "valor é obrigatório", "valor")
			return
		}
		criado, err := useCase.Execute(request.Context(), domain.NovoServicoInput{
			Nome: input.Nome, Descricao: input.Descricao, Valor: input.Valor.String(),
			TempoEstimadoMinutos: input.TempoEstimadoMinutos,
		}, seguranca.UsuarioID(request.Context()))
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(servicoResponse{
			ID: criado.ID, Codigo: criado.Codigo, Nome: criado.Nome, Descricao: criado.Descricao,
			Valor: json.Number(criado.Valor), TempoEstimadoMinutos: criado.TempoEstimadoMinutos,
			Ativo: criado.Ativo, Version: criado.Version,
			DataCriacao: criado.DataCriacao.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func writeError(writer http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrServicoDuplicado) {
		problem(writer, http.StatusConflict, "Conflito", err.Error(), "nome")
		return
	}
	if errors.Is(err, domain.ErrNomeObrigatorio) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "nome")
		return
	}
	if errors.Is(err, domain.ErrValorInvalido) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "valor")
		return
	}
	if errors.Is(err, domain.ErrTempoEstimadoInvalido) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "tempoEstimadoMinutos")
		return
	}
	problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao cadastrar serviço", "")
}

func problem(writer http.ResponseWriter, status int, title, detail, campo string) {
	response := sharedhttp.Problem{
		Type: "https://api.oficina-mecanica.dev/errors/servicos", Title: title, Status: status, Detail: detail,
	}
	if campo != "" {
		response.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, response)
}
