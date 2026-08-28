package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type finalizarRequest struct {
	Observacoes string `json:"observacoes"`
}

type finalizarResponse struct {
	OrdemServicoID  string `json:"ordemServicoId"`
	Status          string `json:"status"`
	DataFinalizacao string `json:"dataFinalizacao"`
	Observacoes     string `json:"observacoes,omitempty"`
}

type itemPendenteResponse struct {
	ItemID     string  `json:"itemId"`
	Codigo     string  `json:"codigo"`
	Quantidade float64 `json:"quantidade"`
}

func NewFinalizarHandler(useCase application.Finalizar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		var input finalizarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}
		resultado, err := useCase.Execute(request.Context(), application.FinalizarInput{OSID: osID, Observacoes: input.Observacoes})
		if err != nil {
			writeFinalizarError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(finalizarResponse{
			OrdemServicoID: resultado.OrdemServicoID, Status: resultado.Status,
			DataFinalizacao: resultado.DataFinalizacao.Format(time.RFC3339), Observacoes: resultado.Observacoes,
		})
	}
}

func writeFinalizarError(writer http.ResponseWriter, err error) {
	var reservasPendentes domain.ErroReservasPendentes
	if errors.As(err, &reservasPendentes) {
		itens := make([]itemPendenteResponse, 0, len(reservasPendentes.Itens))
		for _, item := range reservasPendentes.Itens {
			itens = append(itens, itemPendenteResponse{ItemID: item.ItemID, Codigo: item.Codigo, Quantidade: item.Quantidade})
		}
		problem := sharedhttp.Problem{
			Type: "https://api.oficina-mecanica.dev/errors/ordem-servico", Title: "Conflito de estado",
			Status: http.StatusConflict, Detail: reservasPendentes.Error(), Erros: itens,
		}
		sharedhttp.WriteProblem(writer, problem)
		return
	}
	switch {
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "osId")
	case errors.Is(err, domain.ErrOSNaoEmExecucao), errors.Is(err, domain.ErrServicosPendentes), errors.Is(err, domain.ErrOrcamentoComplementarPendente):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao finalizar servico", "")
	}
}
