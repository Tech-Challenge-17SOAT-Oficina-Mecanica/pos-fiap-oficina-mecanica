package servico

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

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

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrServicoDuplicado):
		problem(w, 409, "Conflito", err.Error(), "nome")
	case errors.Is(err, application.ErrServicoNaoEncontrado):
		problem(w, 404, "Recurso não encontrado", err.Error(), "")
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
