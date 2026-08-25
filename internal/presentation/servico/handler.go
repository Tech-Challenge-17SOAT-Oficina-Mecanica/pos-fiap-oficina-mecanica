package servico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

// ── DTOs ─────────────────────────────────────────────────────────────────────

type cadastrarRequest struct {
	Nome                 string  `json:"nome"`
	Descricao            string  `json:"descricao"`
	Valor                float64 `json:"valor"`
	TempoEstimadoMinutos int     `json:"tempoEstimadoMinutos"`
}

type servicoResponse struct {
	ID                   string  `json:"id"`
	Codigo               string  `json:"codigo"`
	Nome                 string  `json:"nome"`
	Descricao            string  `json:"descricao,omitempty"`
	Valor                float64 `json:"valor"`
	TempoEstimadoMinutos int     `json:"tempoEstimadoMinutos"`
	Ativo                bool    `json:"ativo"`
	Version              int     `json:"version"`
	DataCriacao          string  `json:"dataCriacao"`
}

type servicoAtualizadoResponse struct {
	ID                   string  `json:"id"`
	Codigo               string  `json:"codigo"`
	Nome                 string  `json:"nome"`
	Descricao            string  `json:"descricao,omitempty"`
	Valor                float64 `json:"valor"`
	TempoEstimadoMinutos int     `json:"tempoEstimadoMinutos"`
	Ativo                bool    `json:"ativo"`
	Version              int     `json:"version"`
	DataAtualizacao      string  `json:"dataAtualizacao"`
}

type servicoSituacaoResponse struct {
	ID              string  `json:"id"`
	Codigo          string  `json:"codigo"`
	Nome            string  `json:"nome"`
	Ativo           bool    `json:"ativo"`
	DataDesativacao *string `json:"dataDesativacao,omitempty"`
	UsuarioDesativ  *string `json:"usuarioDesativacao,omitempty"`
}

type servicoResumoResponse struct {
	ID                   string  `json:"id"`
	Codigo               string  `json:"codigo"`
	Nome                 string  `json:"nome"`
	Descricao            string  `json:"descricao,omitempty"`
	Valor                float64 `json:"valor"`
	TempoEstimadoMinutos int     `json:"tempoEstimadoMinutos"`
	Ativo                bool    `json:"ativo"`
}

type servicosResponse struct {
	Data           []servicoResumoResponse `json:"data"`
	Pagina         int                     `json:"pagina"`
	Tamanho        int                     `json:"tamanho"`
	TotalElementos int                     `json:"totalElementos"`
	TotalPaginas   int                     `json:"totalPaginas"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func NewCadastrarHandler(uc application.CadastrarServico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input cadastrarRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&input); err != nil || dec.Decode(&struct{}{}) != io.EOF {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}

		servico, err := uc.Execute(r.Context(), domain.NovoCadastroInput{
			Nome:                 input.Nome,
			Descricao:            input.Descricao,
			Valor:                input.Valor,
			TempoEstimadoMinutos: input.TempoEstimadoMinutos,
		})
		if err != nil {
			if errors.Is(err, domain.ErrNomeObrigatorio) || errors.Is(err, domain.ErrValorInvalido) || errors.Is(err, domain.ErrTempoEstimadoObrigatorio) {
				writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "servico")
				return
			}
			if errors.Is(err, application.ErrNomeDuplicado) {
				writeProblem(w, http.StatusConflict, "Conflito de estado", err.Error(), "nome")
				return
			}
			writeProblem(w, http.StatusInternalServerError, "Erro interno", "erro ao cadastrar serviço", "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(newServicoCadastroResponse(servico))
	}
}

func NewListarHandler(uc application.ConsultarServicos) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pagina, err := intQuery(r, "pagina", 0)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "página inválida", "pagina")
			return
		}
		tamanho, err := intQuery(r, "tamanho", application.TamanhoPaginaPadrao)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "tamanho inválido", "tamanho")
			return
		}
		incluirInativos, err := boolQuery(r, "incluirInativos", false)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "incluirInativos inválido", "incluirInativos")
			return
		}

		resultado, err := uc.Execute(r.Context(), application.FiltrosConsulta{
			Nome:            r.URL.Query().Get("nome"),
			IncluirInativos: incluirInativos,
			Pagina:          pagina,
			Tamanho:         tamanho,
		})
		if err != nil {
			if errors.Is(err, application.ErrConsultaInvalida) {
				writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "consulta")
				return
			}
			writeProblem(w, http.StatusInternalServerError, "Erro interno", "erro ao consultar serviços", "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(servicosResponse{
			Data:           servicosResumo(resultado.Data),
			Pagina:         resultado.Pagina,
			Tamanho:        resultado.Tamanho,
			TotalElementos: resultado.TotalElementos,
			TotalPaginas:   resultado.TotalPaginas,
		})
	}
}

func NewBuscarPorIDHandler(uc application.ConsultarServicoPorID) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicoID := r.PathValue("servicoId")
		if !validation.IsUUID(servicoID) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "servicoId inválido", "servicoId")
			return
		}
		servico, err := uc.Execute(r.Context(), servicoID)
		if err != nil {
			if errors.Is(err, application.ErrServicoNaoEncontrado) {
				writeProblem(w, http.StatusNotFound, "Serviço não encontrado", err.Error(), "servicoId")
				return
			}
			writeProblem(w, http.StatusInternalServerError, "Erro interno", "erro ao consultar serviço", "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(newServicoCadastroResponse(servico))
	}
}

func NewAtualizarHandler(uc application.AtualizarServico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicoID := r.PathValue("servicoId")
		if !validation.IsUUID(servicoID) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "servicoId inválido", "servicoId")
			return
		}
		version, ok, err := ifMatchVersion(r.Header.Get("If-Match"))
		if !ok {
			writeProblem(w, http.StatusPreconditionRequired, "Pré-condição obrigatória", "If-Match obrigatório", "If-Match")
			return
		}
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "If-Match inválido", "If-Match")
			return
		}

		input, err := decodePatchRequest(r)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "servico")
			return
		}

		usuarioID := segurancaPresentation.UsuarioID(r.Context())
		servico, err := uc.Execute(r.Context(), servicoID, input, version, usuarioID)
		if err != nil {
			if errors.Is(err, application.ErrServicoNaoEncontrado) {
				writeProblem(w, http.StatusNotFound, "Serviço não encontrado", err.Error(), "servicoId")
				return
			}
			if errors.Is(err, application.ErrServicoInativo) {
				writeProblem(w, http.StatusConflict, "Serviço inativo", err.Error(), "servicoId")
				return
			}
			if errors.Is(err, application.ErrVersaoDivergente) {
				writeProblem(w, http.StatusPreconditionFailed, "Versão divergente", err.Error(), "If-Match")
				return
			}
			if errors.Is(err, application.ErrNomeDuplicado) {
				writeProblem(w, http.StatusConflict, "Conflito de estado", err.Error(), "nome")
				return
			}
			if errors.Is(err, domain.ErrNomeObrigatorio) || errors.Is(err, domain.ErrValorInvalido) || errors.Is(err, domain.ErrTempoEstimadoObrigatorio) || errors.Is(err, application.ErrAtualizacaoInvalida) {
				writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "servico")
				return
			}
			writeProblem(w, http.StatusInternalServerError, "Erro interno", "erro ao atualizar serviço", "")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(newServicoAtualizadoResponse(servico))
	}
}

func NewDesativarHandler(uc application.DesativarServico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicoID := r.PathValue("servicoId")
		if !validation.IsUUID(servicoID) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "servicoId inválido", "servicoId")
			return
		}
		usuarioID := segurancaPresentation.UsuarioID(r.Context())
		servico, err := uc.Execute(r.Context(), servicoID, usuarioID)
		if err != nil {
			handleSituacaoError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(newServicoSituacaoResponse(servico))
	}
}

func NewReativarHandler(uc application.ReativarServico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicoID := r.PathValue("servicoId")
		if !validation.IsUUID(servicoID) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "servicoId inválido", "servicoId")
			return
		}
		usuarioID := segurancaPresentation.UsuarioID(r.Context())
		servico, err := uc.Execute(r.Context(), servicoID, usuarioID)
		if err != nil {
			handleSituacaoError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(newServicoSituacaoResponse(servico))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func handleSituacaoError(w http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrServicoNaoEncontrado) {
		writeProblem(w, http.StatusNotFound, "Serviço não encontrado", err.Error(), "servicoId")
		return
	}
	if errors.Is(err, application.ErrServicoJaInativo) || errors.Is(err, application.ErrServicoJaAtivo) || errors.Is(err, application.ErrNomeAtivoDuplicado) {
		writeProblem(w, http.StatusConflict, "Conflito de estado", err.Error(), "servicoId")
		return
	}
	writeProblem(w, http.StatusInternalServerError, "Erro interno", "erro ao alterar situação do serviço", "")
}

func decodePatchRequest(r *http.Request) (domain.AtualizacaoInput, error) {
	var raw map[string]json.RawMessage
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil || dec.Decode(&struct{}{}) != io.EOF {
		return domain.AtualizacaoInput{}, errors.New("corpo da requisição inválido")
	}
	immutable := map[string]bool{"id": true, "codigo": true, "dataCriacao": true, "ativo": true}
	allowed := map[string]bool{"nome": true, "descricao": true, "valor": true, "tempoEstimadoMinutos": true}
	for field := range raw {
		if immutable[field] {
			return domain.AtualizacaoInput{}, errors.New(field + " não pode ser alterado")
		}
		if !allowed[field] {
			return domain.AtualizacaoInput{}, errors.New("campo desconhecido: " + field)
		}
	}

	var input domain.AtualizacaoInput
	if v, ok := raw["nome"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return domain.AtualizacaoInput{}, errors.New("nome inválido")
		}
		input.Nome = &s
	}
	if v, ok := raw["descricao"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return domain.AtualizacaoInput{}, errors.New("descricao inválida")
		}
		input.Descricao = &s
	}
	if v, ok := raw["valor"]; ok {
		var f float64
		if err := json.Unmarshal(v, &f); err != nil {
			return domain.AtualizacaoInput{}, errors.New("valor inválido")
		}
		input.Valor = &f
	}
	if v, ok := raw["tempoEstimadoMinutos"]; ok {
		var i int
		if err := json.Unmarshal(v, &i); err != nil {
			return domain.AtualizacaoInput{}, errors.New("tempoEstimadoMinutos inválido")
		}
		input.TempoEstimadoMinutos = &i
	}
	return input, nil
}

func newServicoCadastroResponse(s domain.Servico) servicoResponse {
	return servicoResponse{
		ID:                   s.ID,
		Codigo:               s.Codigo,
		Nome:                 s.Nome,
		Descricao:            s.Descricao,
		Valor:                s.Valor,
		TempoEstimadoMinutos: s.TempoEstimadoMinutos,
		Ativo:                s.Ativo,
		Version:              s.Version,
		DataCriacao:          s.DataCriacao.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func newServicoAtualizadoResponse(s domain.Servico) servicoAtualizadoResponse {
	var dataAtualizacao string
	if s.DataAtualizacao != nil {
		dataAtualizacao = s.DataAtualizacao.Format("2006-01-02T15:04:05Z07:00")
	}
	return servicoAtualizadoResponse{
		ID:                   s.ID,
		Codigo:               s.Codigo,
		Nome:                 s.Nome,
		Descricao:            s.Descricao,
		Valor:                s.Valor,
		TempoEstimadoMinutos: s.TempoEstimadoMinutos,
		Ativo:                s.Ativo,
		Version:              s.Version,
		DataAtualizacao:      dataAtualizacao,
	}
}

func newServicoSituacaoResponse(s domain.Servico) servicoSituacaoResponse {
	var dataDesativ *string
	if s.DataDesativacao != nil {
		formatted := s.DataDesativacao.Format("2006-01-02T15:04:05Z07:00")
		dataDesativ = &formatted
	}
	var usuarioDesativ *string
	if s.UsuarioDesativacao != "" {
		usuarioDesativ = &s.UsuarioDesativacao
	}
	return servicoSituacaoResponse{
		ID:              s.ID,
		Codigo:          s.Codigo,
		Nome:            s.Nome,
		Ativo:           s.Ativo,
		DataDesativacao: dataDesativ,
		UsuarioDesativ:  usuarioDesativ,
	}
}

func servicosResumo(servicos []domain.Servico) []servicoResumoResponse {
	result := make([]servicoResumoResponse, 0, len(servicos))
	for _, s := range servicos {
		result = append(result, servicoResumoResponse{
			ID:                   s.ID,
			Codigo:               s.Codigo,
			Nome:                 s.Nome,
			Descricao:            s.Descricao,
			Valor:                s.Valor,
			TempoEstimadoMinutos: s.TempoEstimadoMinutos,
			Ativo:                s.Ativo,
		})
	}
	return result
}

func ifMatchVersion(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false, nil
	}
	v, err := strconv.Atoi(strings.Trim(value, `"`))
	return v, true, err
}

func intQuery(r *http.Request, name string, fallback int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func boolQuery(r *http.Request, name string, fallback bool) (bool, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseBool(value)
}

func writeProblem(w http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{
		Type:   "https://api.oficina-mecanica.dev/errors/servico",
		Title:  title,
		Status: status,
		Detail: detail,
	}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(w, problem)
}
