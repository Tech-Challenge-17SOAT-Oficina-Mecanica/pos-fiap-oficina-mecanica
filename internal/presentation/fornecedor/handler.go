package fornecedor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
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
