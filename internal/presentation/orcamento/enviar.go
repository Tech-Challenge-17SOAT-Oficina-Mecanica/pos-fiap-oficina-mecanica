package orcamento

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	orcamentoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

type envioResponse struct {
	OrcamentoID        string    `json:"orcamentoId"`
	OrdemServicoID     string    `json:"ordemServicoId"`
	StatusOrdemServico string    `json:"statusOrdemServico"`
	EnviadoEm          time.Time `json:"enviadoEm"`
	NotificacaoEnviada bool      `json:"notificacaoEnviada"`
}

// NewEnviarHandler coloca o orcamento sob decisao do cliente, movendo a OS para
// AGUARDANDO_APROVACAO — a transicao entre calcular e aprovar.
func NewEnviarHandler(useCase orcamento.Enviar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resultado, err := useCase.Execute(request.Context(),
			request.PathValue("orcamentoId"),
			segurancaPresentation.UsuarioID(request.Context()))
		if err != nil {
			responderErroEnvio(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(envioResponse{
			OrcamentoID:        resultado.OrcamentoID,
			OrdemServicoID:     resultado.OrdemServicoID,
			StatusOrdemServico: resultado.StatusOrdemServico,
			EnviadoEm:          resultado.EnviadoEm,
			NotificacaoEnviada: resultado.NotificacaoEnviada,
		})
	}
}

func responderErroEnvio(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, orcamento.ErrIdentificadorInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "orcamentoId")
	case errors.Is(err, orcamento.ErrOrcamentoNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Não encontrado", err.Error(), "")
	case errors.Is(err, orcamentoDominio.ErrOrcamentoNaoCalculado):
		writeProblem(writer, http.StatusConflict, "Conflito", err.Error(), "")
	case errors.Is(err, orcamentoDominio.ErrOrcamentoJaEnviado),
		errors.Is(err, orcamentoDominio.ErrOSNaoPermiteEnvio),
		errors.Is(err, orcamentoDominio.ErrStatusNaoCalculavel),
		errors.Is(err, orcamentoDominio.ErrSemItens):
		writeProblem(writer, http.StatusConflict, "Conflito", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "falha ao enviar o orçamento", "")
	}
}
