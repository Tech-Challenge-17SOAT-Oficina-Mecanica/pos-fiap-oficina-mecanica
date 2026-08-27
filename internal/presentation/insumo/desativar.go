package insumo

import (
	"encoding/json"
	"errors"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type insumoDesativadoResponse struct {
	ID                 string `json:"id"`
	Codigo             string `json:"codigo"`
	Nome               string `json:"nome"`
	UnidadeMedida      string `json:"unidadeMedida"`
	Ativo              bool   `json:"ativo"`
	DataDesativacao    string `json:"dataDesativacao"`
	UsuarioDesativacao string `json:"usuarioDesativacao"`
}

func NewDesativarHandler(useCase application.DesativarInsumo) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		insumoID := request.PathValue("insumoId")
		if !validation.IsUUID(insumoID) {
			problema(writer, http.StatusBadRequest, "Dados inválidos", "insumoId inválido", "insumoId")
			return
		}
		item, err := useCase.Execute(request.Context(), insumoID, segurancaPresentation.UsuarioID(request.Context()))
		if err != nil {
			responderErroDesativacao(writer, err)
			return
		}
		usuario := ""
		if item.UsuarioDesativacao != nil {
			usuario = *item.UsuarioDesativacao
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(insumoDesativadoResponse{
			ID: item.ID, Codigo: item.Codigo, Nome: item.Nome, UnidadeMedida: item.UnidadeMedida,
			Ativo: item.Ativo, DataDesativacao: item.DataDesativacao.Format("2006-01-02T15:04:05Z07:00"), UsuarioDesativacao: usuario,
		})
	}
}

func responderErroDesativacao(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrInsumoNaoEncontrado):
		problema(writer, http.StatusNotFound, "Recurso não encontrado", err.Error(), "insumoId")
	case errors.Is(err, domain.ErrInsumoJaInativo):
		problema(writer, http.StatusConflict, "Conflito de estado", err.Error(), "insumo")
	default:
		var emUso application.InsumoEmUsoError
		if errors.As(err, &emUso) {
			erros := make([]sharedhttp.FieldError, 0, len(emUso.OrdensServico))
			for _, ordemID := range emUso.OrdensServico {
				erros = append(erros, sharedhttp.FieldError{Campo: "ordemServicoId", Mensagem: ordemID})
			}
			sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/insumo-em-uso", Title: "Insumo em uso", Status: http.StatusConflict, Detail: err.Error(), Erros: erros})
			return
		}
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao desativar insumo", "")
	}
}
