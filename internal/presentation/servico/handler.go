package servico

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type CadastrarUseCase interface {
	Execute(context.Context, domain.NovoServicoInput, string) (domain.Servico, error)
}
type ConsultarUseCase interface {
	Listar(context.Context, application.Filtros) (application.Pagina, error)
	PorID(context.Context, string) (domain.Servico, error)
}

type AtualizarUseCase interface {
	Execute(context.Context, string, int, domain.Atualizacao, string) (domain.Servico, error)
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
type servicoListaResponse struct {
	ID                   string      `json:"id"`
	Codigo               string      `json:"codigo"`
	Nome                 string      `json:"nome"`
	Descricao            string      `json:"descricao,omitempty"`
	Valor                json.Number `json:"valor"`
	TempoEstimadoMinutos int         `json:"tempoEstimadoMinutos"`
	Ativo                bool        `json:"ativo"`
}
type listaResponse struct {
	Data           []servicoListaResponse `json:"data"`
	Pagina         int                    `json:"pagina"`
	Tamanho        int                    `json:"tamanho"`
	TotalElementos int                    `json:"totalElementos"`
	TotalPaginas   int                    `json:"totalPaginas"`
}
type servicoDetalheResponse struct {
	servicoListaResponse
	Version int `json:"version"`
}

type atualizarRequest struct {
	Nome                 *string      `json:"nome"`
	Descricao            *string      `json:"descricao"`
	Valor                *json.Number `json:"valor"`
	TempoEstimadoMinutos *int         `json:"tempoEstimadoMinutos"`
}

type servicoAtualizadoResponse struct {
	servicoDetalheResponse
	DataAtualizacao string `json:"dataAtualizacao"`
}

func NewCadastrarHandler(useCase CadastrarUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input cadastrarRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			problem(w, 400, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		if input.Valor == nil {
			problem(w, 400, "Dados inválidos", "valor é obrigatório", "valor")
			return
		}
		criado, err := useCase.Execute(r.Context(), domain.NovoServicoInput{Nome: input.Nome, Descricao: input.Descricao, Valor: input.Valor.String(), TempoEstimadoMinutos: input.TempoEstimadoMinutos}, seguranca.UsuarioID(r.Context()))
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(servicoResponse{ID: criado.ID, Codigo: criado.Codigo, Nome: criado.Nome, Descricao: criado.Descricao, Valor: json.Number(criado.Valor), TempoEstimadoMinutos: criado.TempoEstimadoMinutos, Ativo: criado.Ativo, Version: criado.Version, DataCriacao: criado.DataCriacao.Format("2006-01-02T15:04:05Z07:00")})
	}
}

func NewListarHandler(useCase ConsultarUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pagina, err := inteiroQuery(r, "pagina", 0)
		if err != nil {
			problem(w, 400, "Dados inválidos", "pagina inválida", "pagina")
			return
		}
		tamanho, err := inteiroQuery(r, "tamanho", 20)
		if err != nil {
			problem(w, 400, "Dados inválidos", "tamanho inválido", "tamanho")
			return
		}
		incluirInativos, err := booleanoQuery(r, "incluirInativos", false)
		if err != nil {
			problem(w, 400, "Dados inválidos", "incluirInativos inválido", "incluirInativos")
			return
		}
		resultado, err := useCase.Listar(r.Context(), application.Filtros{Nome: r.URL.Query().Get("nome"), IncluirInativos: incluirInativos, Pagina: pagina, Tamanho: tamanho})
		if err != nil {
			writeError(w, err)
			return
		}
		data := make([]servicoListaResponse, 0, len(resultado.Servicos))
		for _, item := range resultado.Servicos {
			data = append(data, toListaResponse(item))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listaResponse{Data: data, Pagina: resultado.Pagina, Tamanho: resultado.Tamanho, TotalElementos: resultado.TotalElementos, TotalPaginas: resultado.TotalPaginas})
	}
}

func NewConsultarHandler(useCase ConsultarUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("servicoId")
		if !validation.IsUUID(id) {
			problem(w, 400, "Dados inválidos", "servicoId inválido", "servicoId")
			return
		}
		item, err := useCase.PorID(r.Context(), id)
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servicoDetalheResponse{servicoListaResponse: toListaResponse(item), Version: item.Version})
	}
}

func NewAtualizarHandler(useCase AtualizarUseCase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("servicoId")
		if !validation.IsUUID(id) {
			problem(w, http.StatusBadRequest, "Dados inválidos", "servicoId inválido", "servicoId")
			return
		}
		ifMatch := strings.TrimSpace(r.Header.Get("If-Match"))
		if ifMatch == "" {
			problem(w, http.StatusPreconditionRequired, "Pré-condição obrigatória", "If-Match não informado", "If-Match")
			return
		}
		version, err := strconv.Atoi(ifMatch)
		if err != nil || version < 1 {
			problem(w, http.StatusBadRequest, "Dados inválidos", "If-Match inválido", "If-Match")
			return
		}
		input, err := decodeAtualizarRequest(r)
		if err != nil {
			problem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "")
			return
		}
		var valor *string
		if input.Valor != nil {
			texto := input.Valor.String()
			valor = &texto
		}
		atualizado, err := useCase.Execute(r.Context(), id, version, domain.Atualizacao{
			Nome: input.Nome, Descricao: input.Descricao, Valor: valor, TempoEstimadoMinutos: input.TempoEstimadoMinutos,
		}, seguranca.UsuarioID(r.Context()))
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servicoAtualizadoResponse{
			servicoDetalheResponse: servicoDetalheResponse{servicoListaResponse: toListaResponse(atualizado), Version: atualizado.Version},
			DataAtualizacao:        atualizado.DataAtualizacao.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func decodeAtualizarRequest(r *http.Request) (atualizarRequest, error) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&raw); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return atualizarRequest{}, errors.New("corpo da requisição inválido")
	}
	if len(raw) == 0 {
		return atualizarRequest{}, domain.ErrAtualizacaoVazia
	}
	permitidos := map[string]bool{"nome": true, "descricao": true, "valor": true, "tempoEstimadoMinutos": true}
	for campo := range raw {
		if campo == "id" || campo == "codigo" || campo == "dataCriacao" || campo == "ativo" {
			return atualizarRequest{}, errors.New(campo + " não pode ser alterado")
		}
		if !permitidos[campo] {
			return atualizarRequest{}, errors.New("campo desconhecido: " + campo)
		}
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return atualizarRequest{}, err
	}
	var input atualizarRequest
	if err := json.Unmarshal(encoded, &input); err != nil {
		return atualizarRequest{}, errors.New("corpo da requisição inválido")
	}
	return input, nil
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrServicoDuplicado):
		problem(w, 409, "Conflito", err.Error(), "nome")
	case errors.Is(err, application.ErrServicoNaoEncontrado):
		problem(w, 404, "Recurso não encontrado", err.Error(), "")
	case errors.Is(err, application.ErrVersaoDivergente):
		problem(w, 412, "Pré-condição falhou", err.Error(), "If-Match")
	case errors.Is(err, application.ErrPaginaInvalida):
		problem(w, 400, "Dados inválidos", err.Error(), "pagina")
	case errors.Is(err, application.ErrTamanhoInvalido):
		problem(w, 400, "Dados inválidos", err.Error(), "tamanho")
	case errors.Is(err, domain.ErrNomeObrigatorio):
		problem(w, 400, "Dados inválidos", err.Error(), "nome")
	case errors.Is(err, domain.ErrValorInvalido):
		problem(w, 400, "Dados inválidos", err.Error(), "valor")
	case errors.Is(err, domain.ErrTempoEstimadoInvalido):
		problem(w, 400, "Dados inválidos", err.Error(), "tempoEstimadoMinutos")
	case errors.Is(err, domain.ErrAtualizacaoVazia):
		problem(w, 400, "Dados inválidos", err.Error(), "")
	default:
		problem(w, 500, "Erro interno", "falha ao processar serviço", "")
	}
}
func inteiroQuery(r *http.Request, nome string, padrao int) (int, error) {
	valor := r.URL.Query().Get(nome)
	if valor == "" {
		return padrao, nil
	}
	return strconv.Atoi(valor)
}
func booleanoQuery(r *http.Request, nome string, padrao bool) (bool, error) {
	valor := r.URL.Query().Get(nome)
	if valor == "" {
		return padrao, nil
	}
	return strconv.ParseBool(valor)
}
func toListaResponse(s domain.Servico) servicoListaResponse {
	return servicoListaResponse{ID: s.ID, Codigo: s.Codigo, Nome: s.Nome, Descricao: s.Descricao, Valor: json.Number(s.Valor), TempoEstimadoMinutos: s.TempoEstimadoMinutos, Ativo: s.Ativo}
}
func problem(w http.ResponseWriter, status int, title, detail, campo string) {
	response := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/servicos", Title: title, Status: status, Detail: detail}
	if campo != "" {
		response.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(w, response)
}
