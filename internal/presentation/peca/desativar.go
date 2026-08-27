package peca

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

type desativacaoResponse struct {
	ID                 string     `json:"id"`
	Codigo             string     `json:"codigo"`
	Nome               string     `json:"nome"`
	Ativo              bool       `json:"ativo"`
	DataDesativacao    *time.Time `json:"dataDesativacao"`
	UsuarioDesativacao *string    `json:"usuarioDesativacao"`
}

func NewDesativarPecaHandler(useCase peca.DesativarPeca) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		usuarioID := segurancaPresentation.UsuarioID(request.Context())

		desativada, err := useCase.Execute(request.Context(), request.PathValue("pecaId"), usuarioID)
		if err != nil {
			responderErroDesativacao(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(desativacaoResponse{
			ID:                 desativada.ID,
			Codigo:             desativada.Codigo,
			Nome:               desativada.Nome,
			Ativo:              desativada.Ativo,
			DataDesativacao:    desativada.DataDesativacao,
			UsuarioDesativacao: desativada.UsuarioDesativacao,
		})
	}
}

func responderErroDesativacao(writer http.ResponseWriter, err error) {
	var (
		saldoReservado peca.ErroSaldoReservado
		emOrcamento    peca.ErroEmOrcamento
	)

	switch {
	case errors.Is(err, peca.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "pecaId")
	case errors.Is(err, peca.ErrNaoEncontrada):
		problema(writer, http.StatusNotFound, "Não encontrado", "peca inexistente", "")
	case errors.Is(err, pecaDomain.ErrJaInativa):
		problema(writer, http.StatusConflict, "Conflito", err.Error(), "")
	case errors.As(err, &saldoReservado):
		problema(writer, http.StatusConflict, "Conflito", saldoReservado.Error(), "")
	case errors.As(err, &emOrcamento):
		problema(writer, http.StatusConflict, "Conflito", emOrcamento.Error(), "")
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao desativar peça", "")
	}
}
