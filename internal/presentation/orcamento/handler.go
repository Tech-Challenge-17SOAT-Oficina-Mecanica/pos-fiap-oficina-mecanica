package orcamento

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const escopoCalcularOrcamento = "orcamentos:escrever"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type CalcularUseCase interface {
	Execute(context.Context, string, string) (json.Number, error)
}

type TokenValidator interface {
	Validar(string) (seguranca.Claims, error)
}

type calcularResponse struct {
	ValorTotalGeral json.Number `json:"valorTotalGeral"`
}

func NewCalcularHandler(useCase CalcularUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		claims, ok := autorizado(writer, request, token)
		if !ok {
			return
		}
		orcamentoID := request.PathValue("orcamentoId")
		if !uuidPattern.MatchString(orcamentoID) {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "orcamentoId inválido")
			return
		}
		total, err := useCase.Execute(request.Context(), orcamentoID, claims.UsuarioID)
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(calcularResponse{ValorTotalGeral: total})
	}
}

func autorizado(writer http.ResponseWriter, request *http.Request, token TokenValidator) (seguranca.Claims, bool) {
	header := request.Header.Get("Authorization")
	raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if raw == "" || raw == header {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return seguranca.Claims{}, false
	}
	claims, err := token.Validar(raw)
	if err != nil {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return seguranca.Claims{}, false
	}
	for _, escopo := range claims.Escopos {
		if escopo == escopoCalcularOrcamento {
			return claims, true
		}
	}
	problem(writer, http.StatusForbidden, "Acesso negado", "usuário sem o escopo "+escopoCalcularOrcamento)
	return seguranca.Claims{}, false
}

func writeError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrOrcamentoNaoEncontrado):
		problem(writer, http.StatusNotFound, "Não encontrado", err.Error())
	case errors.Is(err, domain.ErrItemInvalido):
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
	case errors.Is(err, domain.ErrStatusInvalido), errors.Is(err, domain.ErrTipoInvalido),
		errors.Is(err, domain.ErrVinculoPrincipalInvalido), errors.Is(err, domain.ErrSemItens),
		errors.Is(err, domain.ErrPrazoIndisponivel), errors.Is(err, domain.ErrTempoServicoIndisponivel),
		errors.Is(err, domain.ErrConfiguracaoInvalida):
		problem(writer, http.StatusConflict, "Orçamento não pode ser calculado", err.Error())
	default:
		problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao calcular orçamento")
	}
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/calculo-orcamento", Title: title, Status: status, Detail: detail})
}
