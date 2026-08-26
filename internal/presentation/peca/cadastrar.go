package peca

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type cadastroRequest struct {
	Nome          string       `json:"nome"`
	Descricao     string       `json:"descricao"`
	CategoriaID   string       `json:"categoriaId"`
	Fabricante    *string      `json:"fabricante"`
	PrecoVenda    *json.Number `json:"precoVenda"`
	EstoqueMinimo *int64       `json:"estoqueMinimo"`
}

func NewCadastrarPecaHandler(useCase peca.CadastrarPeca) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var corpo cadastroRequest
		if err := json.NewDecoder(request.Body).Decode(&corpo); err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}

		var precoVenda *string
		if corpo.PrecoVenda != nil {
			preco := corpo.PrecoVenda.String()
			precoVenda = &preco
		}

		cadastro, err := pecaDomain.NovoCadastro(
			corpo.Nome, corpo.Descricao, corpo.CategoriaID,
			corpo.Fabricante, precoVenda, corpo.EstoqueMinimo,
		)
		if err != nil {
			problemaDeErro(writer, http.StatusBadRequest, "Dados inválidos", err)
			return
		}

		cadastrada, err := useCase.Execute(request.Context(), cadastro)
		if err != nil {
			responderErroCadastro(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(montarResponse(cadastrada, nil))
	}
}

func responderErroCadastro(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, peca.ErrCategoriaInvalida),
		errors.Is(err, peca.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "categoriaId")
	case errors.Is(err, peca.ErrDescricaoDuplicada):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "descricao")
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao cadastrar peça", "")
	}
}
