package insumo

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	insumoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type cadastroRequest struct {
	Nome          string       `json:"nome"`
	Descricao     string       `json:"descricao"`
	CategoriaID   string       `json:"categoriaId"`
	UnidadeMedida string       `json:"unidadeMedida"`
	CustoUnitario *json.Number `json:"custoUnitario"`
	EstoqueMinimo *json.Number `json:"estoqueMinimo"`
}

type insumoResponse struct {
	ID             string       `json:"id"`
	Codigo         string       `json:"codigo"`
	Tipo           string       `json:"tipo"`
	Nome           string       `json:"nome"`
	Descricao      string       `json:"descricao"`
	CategoriaID    string       `json:"categoriaId"`
	Categoria      string       `json:"categoria"`
	UnidadeMedida  string       `json:"unidadeMedida"`
	CustoUnitario  *json.Number `json:"custoUnitario"`
	SaldoFisico    json.Number  `json:"saldoFisico"`
	SaldoReservado json.Number  `json:"saldoReservado"`
	EstoqueMinimo  json.Number  `json:"estoqueMinimo"`
	Ativo          bool         `json:"ativo"`
	Version        int          `json:"version"`
	DataCriacao    *time.Time   `json:"dataCriacao,omitempty"`
}

func NewCadastrarInsumoHandler(useCase insumo.CadastrarInsumo) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var corpo cadastroRequest
		if err := json.NewDecoder(request.Body).Decode(&corpo); err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido")
			return
		}

		cadastro, err := insumoDomain.NovoCadastro(
			corpo.Nome, corpo.Descricao, corpo.CategoriaID, corpo.UnidadeMedida,
			texto(corpo.CustoUnitario), texto(corpo.EstoqueMinimo),
		)
		if err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
			return
		}

		cadastrado, err := useCase.Execute(request.Context(), cadastro)
		if err != nil {
			responderErroCadastro(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(montarResponse(cadastrado))
	}
}

func montarResponse(cadastrado insumoDomain.Insumo) insumoResponse {
	resposta := insumoResponse{
		ID:             cadastrado.ID,
		Codigo:         cadastrado.Codigo,
		Tipo:           insumoDomain.Tipo,
		Nome:           cadastrado.Nome,
		Descricao:      cadastrado.Descricao,
		CategoriaID:    cadastrado.CategoriaID,
		Categoria:      cadastrado.Categoria,
		UnidadeMedida:  cadastrado.UnidadeMedida,
		SaldoFisico:    json.Number(cadastrado.SaldoFisico),
		SaldoReservado: json.Number(cadastrado.SaldoReservado),
		EstoqueMinimo:  json.Number(cadastrado.EstoqueMinimo),
		Ativo:          cadastrado.Ativo,
		Version:        cadastrado.Version,
		DataCriacao:    cadastrado.DataCriacao,
	}
	if cadastrado.CustoUnitario != nil {
		custo := json.Number(*cadastrado.CustoUnitario)
		resposta.CustoUnitario = &custo
	}
	return resposta
}

func texto(numero *json.Number) *string {
	if numero == nil {
		return nil
	}
	valor := numero.String()
	return &valor
}

func responderErroCadastro(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, insumo.ErrCategoriaInvalida),
		errors.Is(err, insumo.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
	case errors.Is(err, insumo.ErrDescricaoDuplicada):
		problema(writer, http.StatusConflict, "Conflito", err.Error())
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao cadastrar insumo")
	}
}

func problema(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{
		Type:   "https://api.oficina-mecanica.dev/errors/estoque",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}
