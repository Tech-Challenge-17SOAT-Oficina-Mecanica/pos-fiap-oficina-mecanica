package mecanico

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type CadastrarUseCase interface {
	Execute(context.Context, domain.NovoMecanicoInput) (domain.Mecanico, error)
}

type AtualizarUseCase interface {
	Execute(context.Context, application.AtualizarInput) (domain.Mecanico, error)
}

type cadastrarRequest struct {
	Nome    string   `json:"nome"`
	Email   string   `json:"email"`
	Senha   string   `json:"senha"`
	Escopos []string `json:"escopos"`
}

type atualizarRequest struct {
	Nome    string   `json:"nome"`
	Email   string   `json:"email"`
	Escopos []string `json:"escopos"`
}

type mecanicoResponse struct {
	ID      string   `json:"id"`
	Nome    string   `json:"nome"`
	Email   string   `json:"email"`
	Ativo   bool     `json:"ativo"`
	Escopos []string `json:"escopos"`
}

type atualizarResponse struct {
	ID      string   `json:"id"`
	Nome    string   `json:"nome"`
	Email   string   `json:"email"`
	Ativo   bool     `json:"ativo"`
	Escopos []string `json:"escopos"`
	Version int      `json:"version"`
}

func NewCadastrarHandler(useCase CadastrarUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input cadastrarRequest
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido")
			return
		}
		mecanico, err := useCase.Execute(request.Context(), domain.NovoMecanicoInput{
			Nome:    input.Nome,
			Email:   input.Email,
			Senha:   input.Senha,
			Escopos: input.Escopos,
		})
		if err != nil {
			writeCadastrarError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(mecanicoResponse{
			ID:      mecanico.ID,
			Nome:    mecanico.Nome,
			Email:   mecanico.Email,
			Ativo:   mecanico.Ativo,
			Escopos: mecanico.Escopos,
		})
	}
}

func NewAtualizarHandler(useCase AtualizarUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		mecanicoID := request.PathValue("mecanicoId")
		if !validation.IsUUID(mecanicoID) {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "mecanicoId inválido")
			return
		}
		ifMatch := strings.TrimSpace(request.Header.Get("If-Match"))
		if ifMatch == "" {
			problem(writer, http.StatusPreconditionRequired, "Pré-condição obrigatória", "If-Match não informado")
			return
		}
		version, err := strconv.Atoi(ifMatch)
		if err != nil || version <= 0 {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "If-Match inválido")
			return
		}
		var input atualizarRequest
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido")
			return
		}
		mecanico, err := useCase.Execute(request.Context(), application.AtualizarInput{
			MecanicoID: mecanicoID,
			Version:    version,
			Dados: domain.AtualizarMecanicoInput{
				Nome:    input.Nome,
				Email:   input.Email,
				Escopos: input.Escopos,
			},
		})
		if err != nil {
			writeAtualizarError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(atualizarResponse{
			ID:      mecanico.ID,
			Nome:    mecanico.Nome,
			Email:   mecanico.Email,
			Ativo:   mecanico.Ativo,
			Escopos: mecanico.Escopos,
			Version: mecanico.Version,
		})
	}
}

func writeCadastrarError(writer http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrEmailDuplicado) {
		problem(writer, http.StatusConflict, "Conflito", err.Error())
		return
	}
	if errMecanicoInvalido(err) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
		return
	}
	problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao cadastrar mecânico")
}

func writeAtualizarError(writer http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrEmailDuplicado) {
		problem(writer, http.StatusConflict, "Conflito", err.Error())
		return
	}
	if errors.Is(err, application.ErrMecanicoNaoEncontrado) {
		problem(writer, http.StatusNotFound, "Não encontrado", err.Error())
		return
	}
	if errors.Is(err, application.ErrVersaoDivergente) {
		problem(writer, http.StatusPreconditionFailed, "Pré-condição falhou", err.Error())
		return
	}
	if errMecanicoInvalido(err) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
		return
	}
	problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao atualizar mecânico")
}

func errMecanicoInvalido(err error) bool {
	return errors.Is(err, domain.ErrNomeObrigatorio) ||
		errors.Is(err, domain.ErrMecanicoIDObrigatorio) ||
		errors.Is(err, domain.ErrEmailObrigatorio) ||
		errors.Is(err, domain.ErrSenhaObrigatoria) ||
		errors.Is(err, domain.ErrSenhaCurta) ||
		errors.Is(err, domain.ErrEscoposObrigatorio) ||
		errors.Is(err, domain.ErrEmailInvalido) ||
		errors.Is(err, domain.ErrEscopoInvalido)
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/mecanicos", Title: title, Status: status, Detail: detail})
}
