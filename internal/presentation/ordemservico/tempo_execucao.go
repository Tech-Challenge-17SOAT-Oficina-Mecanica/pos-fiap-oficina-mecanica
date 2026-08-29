package ordemservico

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type tempoExecucaoResponse struct {
	OrdemServicoID       string `json:"ordemServicoId"`
	DataInicioExecucao   string `json:"dataInicioExecucao"`
	DataFinalizacao      string `json:"dataFinalizacao"`
	TempoExecucaoMinutos int    `json:"tempoExecucaoMinutos"`
}

type temposExecucaoResponse struct {
	TempoMedioExecucaoMinutos int                     `json:"tempoMedioExecucaoMinutos"`
	Data                      []tempoExecucaoResponse `json:"data"`
	Pagina                    int                     `json:"pagina"`
	Tamanho                   int                     `json:"tamanho"`
	TotalElementos            int                     `json:"totalElementos"`
	TotalPaginas              int                     `json:"totalPaginas"`
}

func NewConsultarTempoExecucaoHandler(useCase application.ConsultarTempoExecucaoDaOS) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		resultado, err := useCase.Execute(request.Context(), osID)
		if err != nil {
			writeTempoExecucaoError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(tempoResponse(resultado))
	}
}

func NewListarTemposExecucaoHandler(useCase application.ConsultarTempoMedioExecucao) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		paginacao, err := sharedhttp.LerPaginacao(request.URL.Query())
		if err != nil {
			writeCampoInvalido(writer, err)
			return
		}
		resultado, err := useCase.Execute(request.Context(), application.ConsultarTempoMedioExecucaoInput{
			DataInicio: request.URL.Query().Get("dataInicio"), DataFim: request.URL.Query().Get("dataFim"),
			Limite: paginacao.Limit(), Deslocamento: paginacao.Offset(),
		})
		if err != nil {
			writeTempoExecucaoError(writer, err)
			return
		}
		itens := make([]tempoExecucaoResponse, 0, len(resultado.Itens))
		for _, item := range resultado.Itens {
			itens = append(itens, tempoResponse(item))
		}
		lista := sharedhttp.NovaLista(itens, paginacao, resultado.TotalElementos)
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(temposExecucaoResponse{
			TempoMedioExecucaoMinutos: resultado.TempoMedioExecucaoMinutos, Data: lista.Data,
			Pagina: lista.Pagina, Tamanho: lista.Tamanho, TotalElementos: lista.TotalElementos, TotalPaginas: lista.TotalPaginas,
		})
	}
}

func tempoResponse(resultado domain.TempoExecucao) tempoExecucaoResponse {
	return tempoExecucaoResponse{
		OrdemServicoID: resultado.OrdemServicoID, DataInicioExecucao: resultado.DataInicioExecucao.Format(time.RFC3339),
		DataFinalizacao: resultado.DataFinalizacao.Format(time.RFC3339), TempoExecucaoMinutos: resultado.TempoExecucaoMinutos,
	}
}

func writeTempoExecucaoError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "osId")
	case errors.Is(err, domain.ErrTempoExecucaoIndisponivel):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "osId")
	case errors.Is(err, application.ErrDataInicioInvalida), errors.Is(err, application.ErrPeriodoInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "dataInicio")
	case errors.Is(err, application.ErrDataFimInvalida):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "dataFim")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao consultar tempo de execucao", "")
	}
}
