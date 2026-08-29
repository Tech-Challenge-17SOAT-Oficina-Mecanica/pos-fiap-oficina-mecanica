package peca

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type atualizacaoRequest struct {
	Nome          string       `json:"nome"`
	Descricao     string       `json:"descricao"`
	CategoriaID   string       `json:"categoriaId"`
	FornecedorID  *string      `json:"fornecedorId"`
	Fabricante    *string      `json:"fabricante"`
	PrecoVenda    *json.Number `json:"precoVenda"`
	EstoqueMinimo *int64       `json:"estoqueMinimo"`
	// Ativo existe só para ser recusado: mudar a situação é responsabilidade do DELETE,
	// onde as validações de saldo reservado e orçamento estão.
	Ativo *bool `json:"ativo"`
}

func NewAtualizarPecaHandler(useCase peca.AtualizarPeca) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		version, err := sharedhttp.LerIfMatch(request.Header.Get("If-Match"))
		if err != nil {
			status := http.StatusBadRequest
			titulo := "Dados inválidos"
			if errors.Is(err, sharedhttp.ErrIfMatchAusente) {
				status = http.StatusPreconditionRequired
				titulo = "Pré-condição obrigatória"
			}
			problemaDeErro(writer, status, titulo, err)
			return
		}

		var corpo atualizacaoRequest
		if err := json.NewDecoder(request.Body).Decode(&corpo); err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}

		var precoVenda *string
		if corpo.PrecoVenda != nil {
			preco := corpo.PrecoVenda.String()
			precoVenda = &preco
		}

		atualizacao, err := pecaDomain.NovaAtualizacao(
			corpo.Nome, corpo.Descricao, corpo.CategoriaID,
			corpo.Fabricante, precoVenda, corpo.EstoqueMinimo, corpo.FornecedorID, corpo.Ativo != nil,
		)
		if err != nil {
			problemaDeErro(writer, http.StatusBadRequest, "Dados inválidos", err)
			return
		}

		atualizada, err := useCase.Execute(request.Context(), request.PathValue("pecaId"), version,
			atualizacao, segurancaPresentation.UsuarioID(request.Context()))
		if err != nil {
			responderErroAtualizacao(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(montarResponse(atualizada, nil))
	}
}

func responderErroAtualizacao(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, peca.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "pecaId")
	case errors.Is(err, peca.ErrCategoriaInvalida):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "categoriaId")
	case errors.Is(err, peca.ErrFornecedorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "fornecedorId")
	case errors.Is(err, peca.ErrNaoEncontrada):
		problema(writer, http.StatusNotFound, "Não encontrado", "peca inexistente", "")
	case errors.Is(err, peca.ErrVersaoDivergente):
		problema(writer, http.StatusPreconditionFailed, "Versão divergente", err.Error(), "If-Match")
	case errors.Is(err, peca.ErrDescricaoDuplicada):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "descricao")
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao atualizar peça", "")
	}
}
