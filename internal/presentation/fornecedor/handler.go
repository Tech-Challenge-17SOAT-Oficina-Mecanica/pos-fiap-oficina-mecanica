package fornecedor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

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
		_ = json.NewEncoder(writer).Encode(fornecedorResponse{
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
		})
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
