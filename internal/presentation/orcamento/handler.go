package orcamento

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	clienteDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const escopoConsultarOrcamento = "orcamentos:ler"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type ConsultarUseCase interface {
	Execute(context.Context, application.ConsultarInput) (application.Resultado, error)
}

type TokenValidator interface {
	Validar(string) (seguranca.Claims, error)
}

type listaResponse struct {
	Data           []domain.Consulta `json:"data"`
	Pagina         int               `json:"pagina"`
	Tamanho        int               `json:"tamanho"`
	TotalElementos int               `json:"totalElementos"`
	TotalPaginas   int               `json:"totalPaginas"`
}

func NewConsultarHandler(useCase ConsultarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !autorizado(writer, request, token) {
			return
		}
		query := request.URL.Query()
		orcamentoID := strings.TrimSpace(query.Get("orcamentoId"))
		if orcamentoID != "" && !uuidPattern.MatchString(orcamentoID) {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "orcamentoId inválido")
			return
		}
		pagina, ok := parametroInteiro(writer, query.Get("pagina"), 0, "pagina")
		if !ok {
			return
		}
		tamanho, ok := parametroInteiro(writer, query.Get("tamanho"), 20, "tamanho")
		if !ok {
			return
		}
		result, err := useCase.Execute(request.Context(), application.ConsultarInput{OrcamentoID: orcamentoID, Documento: query.Get("documento"), Pagina: pagina, Tamanho: tamanho})
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if result.Consulta != nil {
			_ = json.NewEncoder(writer).Encode(result.Consulta)
			return
		}
		_ = json.NewEncoder(writer).Encode(listaResponse{Data: result.Data, Pagina: result.Pagina, Tamanho: result.Tamanho, TotalElementos: result.TotalElementos, TotalPaginas: result.TotalPaginas})
	}
}

func parametroInteiro(writer http.ResponseWriter, value string, padrao int, nome string) (int, bool) {
	if value == "" {
		return padrao, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		problem(writer, http.StatusBadRequest, "Dados inválidos", nome+" inválida")
		return 0, false
	}
	return parsed, true
}

func autorizado(writer http.ResponseWriter, request *http.Request, token TokenValidator) bool {
	header := request.Header.Get("Authorization")
	raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if raw == "" || raw == header {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return false
	}
	claims, err := token.Validar(raw)
	if err != nil {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return false
	}
	for _, escopo := range claims.Escopos {
		if escopo == escopoConsultarOrcamento {
			return true
		}
	}
	problem(writer, http.StatusForbidden, "Acesso negado", "usuário sem o escopo "+escopoConsultarOrcamento)
	return false
}

func writeError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrCriterioObrigatorio), errors.Is(err, application.ErrPaginacaoInvalida), errors.Is(err, clienteDomain.ErrDocumentoInvalido), errors.Is(err, clienteDomain.ErrDocumentoObrigatorio):
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
	case errors.Is(err, application.ErrOrcamentoNaoEncontrado), errors.Is(err, application.ErrClienteNaoEncontrado):
		problem(writer, http.StatusNotFound, "Não encontrado", err.Error())
	default:
		problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao consultar orçamento")
	}
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/orcamentos", Title: title, Status: status, Detail: detail})
}
