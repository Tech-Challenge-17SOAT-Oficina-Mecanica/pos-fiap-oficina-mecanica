package mecanico

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const escopoCadastrarMecanico = "mecanicos:escrever"

type CadastrarUseCase interface {
	Execute(context.Context, domain.NovoMecanicoInput) (domain.Mecanico, error)
}

type TokenValidator interface {
	Validar(string) (seguranca.Claims, error)
}

type cadastrarRequest struct {
	Nome    string   `json:"nome"`
	Email   string   `json:"email"`
	Senha   string   `json:"senha"`
	Escopos []string `json:"escopos"`
}

type mecanicoResponse struct {
	ID      string   `json:"id"`
	Nome    string   `json:"nome"`
	Email   string   `json:"email"`
	Ativo   bool     `json:"ativo"`
	Escopos []string `json:"escopos"`
}

func NewCadastrarHandler(useCase CadastrarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if status, ok := autorizado(request, token, escopoCadastrarMecanico); !ok {
			if status == http.StatusUnauthorized {
				problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
				return
			}
			problem(writer, http.StatusForbidden, "Acesso negado", "usuário sem o escopo "+escopoCadastrarMecanico)
			return
		}
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

func autorizado(request *http.Request, token TokenValidator, escopoExigido string) (int, bool) {
	header := request.Header.Get("Authorization")
	raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if raw == "" || raw == header {
		return http.StatusUnauthorized, false
	}
	claims, err := token.Validar(raw)
	if err != nil {
		return http.StatusUnauthorized, false
	}
	for _, escopo := range claims.Escopos {
		if escopo == escopoExigido {
			return http.StatusOK, true
		}
	}
	return http.StatusForbidden, false
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

func errMecanicoInvalido(err error) bool {
	return errors.Is(err, domain.ErrNomeObrigatorio) ||
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
