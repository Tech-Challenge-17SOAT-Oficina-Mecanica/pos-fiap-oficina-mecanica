package fornecedor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type cadastrarRequest struct {
	RazaoSocial      string `json:"razaoSocial"`
	NomeFantasia     string `json:"nomeFantasia"`
	Documento        string `json:"documento"`
	TipoDocumento    string `json:"tipoDocumento"`
	Telefone         string `json:"telefone"`
	Email            string `json:"email"`
	PrazoEntregaDias *int   `json:"prazoEntregaDias"`
}

type atualizarRequest struct {
	RazaoSocial      string `json:"razaoSocial"`
	NomeFantasia     string `json:"nomeFantasia"`
	Telefone         string `json:"telefone"`
	Email            string `json:"email"`
	PrazoEntregaDias *int   `json:"prazoEntregaDias"`
}

type fornecedorResponse struct {
	ID               string `json:"id"`
	RazaoSocial      string `json:"razaoSocial"`
	NomeFantasia     string `json:"nomeFantasia,omitempty"`
	Documento        string `json:"documento"`
	TipoDocumento    string `json:"tipoDocumento"`
	Telefone         string `json:"telefone,omitempty"`
	Email            string `json:"email,omitempty"`
	PrazoEntregaDias int    `json:"prazoEntregaDias"`
	Ativo            bool   `json:"ativo"`
	Version          int    `json:"version"`
	DataCriacao      string `json:"dataCriacao"`
}

type fornecedorAtualizadoResponse struct {
	ID               string `json:"id"`
	RazaoSocial      string `json:"razaoSocial"`
	NomeFantasia     string `json:"nomeFantasia,omitempty"`
	Documento        string `json:"documento"`
	TipoDocumento    string `json:"tipoDocumento"`
	Telefone         string `json:"telefone,omitempty"`
	Email            string `json:"email,omitempty"`
	PrazoEntregaDias int    `json:"prazoEntregaDias"`
	Ativo            bool   `json:"ativo"`
	Version          int    `json:"version"`
	DataAtualizacao  string `json:"dataAtualizacao"`
}

type fornecedorSituacaoResponse struct {
	ID           string  `json:"id"`
	RazaoSocial  string  `json:"razaoSocial"`
	Documento    string  `json:"documento"`
	Ativo        bool    `json:"ativo"`
	InativadoEm  *string `json:"inativadoEm"`
	InativadoPor *string `json:"inativadoPor"`
}

type fornecedorResumoResponse struct {
	ID               string `json:"id"`
	RazaoSocial      string `json:"razaoSocial"`
	NomeFantasia     string `json:"nomeFantasia,omitempty"`
	Documento        string `json:"documento"`
	TipoDocumento    string `json:"tipoDocumento"`
	Telefone         string `json:"telefone,omitempty"`
	Email            string `json:"email,omitempty"`
	PrazoEntregaDias int    `json:"prazoEntregaDias"`
	Ativo            bool   `json:"ativo"`
}

type fornecedoresResponse struct {
	Data           []fornecedorResumoResponse `json:"data"`
	Pagina         int                        `json:"pagina"`
	Tamanho        int                        `json:"tamanho"`
	TotalElementos int                        `json:"totalElementos"`
	TotalPaginas   int                        `json:"totalPaginas"`
}

func NewCadastrarHandler(useCase application.Cadastrar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input cadastrarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}

		cadastro, err := domain.NovoCadastro(
			input.RazaoSocial,
			input.NomeFantasia,
			input.Documento,
			input.TipoDocumento,
			input.Telefone,
			input.Email,
			input.PrazoEntregaDias,
		)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "fornecedor")
			return
		}

		fornecedor, err := useCase.Execute(request.Context(), cadastro)
		if err != nil {
			if errors.Is(err, application.ErrDocumentoDuplicado) {
				writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "documento")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao cadastrar fornecedor", "")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(newFornecedorResponse(fornecedor))
	}
}

func NewListarHandler(useCase application.ConsultarFornecedores) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		pagina, err := intQuery(request, "pagina", 0)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "pagina invalida", "pagina")
			return
		}
		tamanho, err := intQuery(request, "tamanho", application.TamanhoPaginaPadrao)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "tamanho invalido", "tamanho")
			return
		}
		incluirInativos, err := boolQuery(request, "incluirInativos", false)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "incluirInativos invalido", "incluirInativos")
			return
		}

		resultado, err := useCase.Execute(request.Context(), application.FiltrosConsulta{
			Nome:            request.URL.Query().Get("nome"),
			Documento:       request.URL.Query().Get("documento"),
			IncluirInativos: incluirInativos,
			Pagina:          pagina,
			Tamanho:         tamanho,
		})
		if err != nil {
			if errors.Is(err, application.ErrConsultaInvalida) {
				writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "consulta")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao consultar fornecedores", "")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(fornecedoresResponse{
			Data:           fornecedoresResumo(resultado.Data),
			Pagina:         resultado.Pagina,
			Tamanho:        resultado.Tamanho,
			TotalElementos: resultado.TotalElementos,
			TotalPaginas:   resultado.TotalPaginas,
		})
	}
}

func NewBuscarPorIDHandler(useCase application.ConsultarFornecedorPorID) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fornecedorID := request.PathValue("fornecedorId")
		if !uuidRegex.MatchString(fornecedorID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "fornecedorId invalido", "fornecedorId")
			return
		}
		fornecedor, err := useCase.Execute(request.Context(), fornecedorID)
		if err != nil {
			if errors.Is(err, application.ErrFornecedorNaoEncontrado) {
				writeProblem(writer, http.StatusNotFound, "Fornecedor nao encontrado", err.Error(), "fornecedorId")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao consultar fornecedor", "")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(newFornecedorResponse(fornecedor))
	}
}

func NewAtualizarHandler(useCase application.AtualizarFornecedor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fornecedorID := request.PathValue("fornecedorId")
		if !uuidRegex.MatchString(fornecedorID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "fornecedorId invalido", "fornecedorId")
			return
		}
		version, ok, err := ifMatchVersion(request.Header.Get("If-Match"))
		if !ok {
			writeProblem(writer, http.StatusPreconditionRequired, "Pre-condicao obrigatoria", "If-Match obrigatorio", "If-Match")
			return
		}
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "If-Match invalido", "If-Match")
			return
		}

		input, err := decodeAtualizarRequest(request)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "fornecedor")
			return
		}
		atualizacao, err := domain.NovaAtualizacao(input.RazaoSocial, input.NomeFantasia, input.Telefone, input.Email, input.PrazoEntregaDias)
		if err != nil {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "fornecedor")
			return
		}
		claims, _ := segurancaPresentation.ClaimsFromContext(request.Context())
		fornecedor, err := useCase.Execute(request.Context(), fornecedorID, atualizacao, version, claims.Subject)
		if err != nil {
			if errors.Is(err, application.ErrFornecedorNaoEncontrado) {
				writeProblem(writer, http.StatusNotFound, "Fornecedor nao encontrado", err.Error(), "fornecedorId")
				return
			}
			if errors.Is(err, application.ErrFornecedorInativo) {
				writeProblem(writer, http.StatusConflict, "Fornecedor inativo", err.Error(), "fornecedorId")
				return
			}
			if errors.Is(err, application.ErrVersaoDivergente) {
				writeProblem(writer, http.StatusPreconditionFailed, "Versao divergente", err.Error(), "If-Match")
				return
			}
			if errors.Is(err, application.ErrAtualizacaoInvalida) {
				writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "fornecedor")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao atualizar fornecedor", "")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(newFornecedorAtualizadoResponse(fornecedor))
	}
}

func NewDesativarHandler(useCase application.DesativarFornecedor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fornecedorID := request.PathValue("fornecedorId")
		if !uuidRegex.MatchString(fornecedorID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "fornecedorId invalido", "fornecedorId")
			return
		}
		claims, _ := segurancaPresentation.ClaimsFromContext(request.Context())
		fornecedor, err := useCase.Execute(request.Context(), fornecedorID, claims.Subject)
		if err != nil {
			handleSituacaoError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(newFornecedorSituacaoResponse(fornecedor))
	}
}

func NewReativarHandler(useCase application.ReativarFornecedor) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		fornecedorID := request.PathValue("fornecedorId")
		if !uuidRegex.MatchString(fornecedorID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "fornecedorId invalido", "fornecedorId")
			return
		}
		claims, _ := segurancaPresentation.ClaimsFromContext(request.Context())
		fornecedor, err := useCase.Execute(request.Context(), fornecedorID, claims.Subject)
		if err != nil {
			handleSituacaoError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(newFornecedorSituacaoResponse(fornecedor))
	}
}

func handleSituacaoError(writer http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrFornecedorNaoEncontrado) {
		writeProblem(writer, http.StatusNotFound, "Fornecedor nao encontrado", err.Error(), "fornecedorId")
		return
	}
	if errors.Is(err, application.ErrFornecedorJaInativo) || errors.Is(err, application.ErrFornecedorJaAtivo) || errors.Is(err, application.ErrFornecedorComPedidoAberto) || errors.Is(err, application.ErrDocumentoAtivoDuplicado) {
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "fornecedorId")
		return
	}
	if errors.Is(err, application.ErrSituacaoInvalida) {
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "fornecedorId")
		return
	}
	writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao alterar situacao do fornecedor", "")
}

func decodeAtualizarRequest(request *http.Request) (atualizarRequest, error) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return atualizarRequest{}, errors.New("corpo da requisicao invalido")
	}
	allowed := map[string]bool{"razaoSocial": true, "nomeFantasia": true, "telefone": true, "email": true, "prazoEntregaDias": true}
	for field := range raw {
		if field == "documento" || field == "ativo" {
			return atualizarRequest{}, errors.New(field + " nao pode ser alterado")
		}
		if !allowed[field] {
			return atualizarRequest{}, errors.New("campo desconhecido: " + field)
		}
	}
	var input atualizarRequest
	encoded, err := json.Marshal(raw)
	if err != nil {
		return atualizarRequest{}, err
	}
	if err := json.Unmarshal(encoded, &input); err != nil {
		return atualizarRequest{}, errors.New("corpo da requisicao invalido")
	}
	return input, nil
}

func ifMatchVersion(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false, nil
	}
	version, err := strconv.Atoi(strings.Trim(value, `"`))
	return version, true, err
}

func intQuery(request *http.Request, name string, fallback int) (int, error) {
	value := request.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func boolQuery(request *http.Request, name string, fallback bool) (bool, error) {
	value := request.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	return strconv.ParseBool(value)
}

func fornecedoresResumo(fornecedores []domain.Fornecedor) []fornecedorResumoResponse {
	response := make([]fornecedorResumoResponse, 0, len(fornecedores))
	for _, fornecedor := range fornecedores {
		response = append(response, fornecedorResumoResponse{
			ID:               fornecedor.ID,
			RazaoSocial:      fornecedor.RazaoSocial,
			NomeFantasia:     fornecedor.NomeFantasia,
			Documento:        fornecedor.Documento,
			TipoDocumento:    fornecedor.TipoDocumento,
			Telefone:         fornecedor.Telefone,
			Email:            fornecedor.Email,
			PrazoEntregaDias: fornecedor.PrazoEntregaDias,
			Ativo:            fornecedor.Ativo,
		})
	}
	return response
}

func newFornecedorResponse(fornecedor domain.Fornecedor) fornecedorResponse {
	return fornecedorResponse{
		ID:               fornecedor.ID,
		RazaoSocial:      fornecedor.RazaoSocial,
		NomeFantasia:     fornecedor.NomeFantasia,
		Documento:        fornecedor.Documento,
		TipoDocumento:    fornecedor.TipoDocumento,
		Telefone:         fornecedor.Telefone,
		Email:            fornecedor.Email,
		PrazoEntregaDias: fornecedor.PrazoEntregaDias,
		Ativo:            fornecedor.Ativo,
		Version:          fornecedor.Version,
		DataCriacao:      fornecedor.CriadoEm.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func newFornecedorAtualizadoResponse(fornecedor domain.Fornecedor) fornecedorAtualizadoResponse {
	return fornecedorAtualizadoResponse{
		ID:               fornecedor.ID,
		RazaoSocial:      fornecedor.RazaoSocial,
		NomeFantasia:     fornecedor.NomeFantasia,
		Documento:        fornecedor.Documento,
		TipoDocumento:    fornecedor.TipoDocumento,
		Telefone:         fornecedor.Telefone,
		Email:            fornecedor.Email,
		PrazoEntregaDias: fornecedor.PrazoEntregaDias,
		Ativo:            fornecedor.Ativo,
		Version:          fornecedor.Version,
		DataAtualizacao:  fornecedor.AtualizadoEm.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func newFornecedorSituacaoResponse(fornecedor domain.Fornecedor) fornecedorSituacaoResponse {
	var inativadoEm *string
	if fornecedor.InativadoEm != nil {
		formatted := fornecedor.InativadoEm.Format("2006-01-02T15:04:05Z07:00")
		inativadoEm = &formatted
	}
	var inativadoPor *string
	if fornecedor.InativadoPor != "" {
		inativadoPor = &fornecedor.InativadoPor
	}
	return fornecedorSituacaoResponse{
		ID:           fornecedor.ID,
		RazaoSocial:  fornecedor.RazaoSocial,
		Documento:    fornecedor.Documento,
		Ativo:        fornecedor.Ativo,
		InativadoEm:  inativadoEm,
		InativadoPor: inativadoPor,
	}
}

func writeProblem(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{
		Type:   "https://api.oficina-mecanica.dev/errors/fornecedor",
		Title:  title,
		Status: status,
		Detail: detail,
	}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}
