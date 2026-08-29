package ordemservico

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type iniciarExecucaoResponse struct {
	OrdemServicoID     string `json:"ordemServicoId"`
	Status             string `json:"status"`
	MecanicoID         string `json:"mecanicoId"`
	DataInicioExecucao string `json:"dataInicioExecucao"`
}

func NewIniciarExecucaoHandler(useCase application.IniciarExecucao) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		if request.Body != nil {
			var body any
			if err := json.NewDecoder(request.Body).Decode(&body); err != io.EOF {
				writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "esta operacao nao aceita corpo na requisicao", "")
				return
			}
		}
		resultado, err := useCase.Execute(request.Context(), application.IniciarExecucaoInput{
			OSID: osID, UsuarioID: seguranca.UsuarioID(request.Context()),
		})
		if err != nil {
			writeIniciarExecucaoError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(iniciarExecucaoResponse{
			OrdemServicoID: resultado.OrdemServicoID, Status: resultado.Status, MecanicoID: resultado.MecanicoID,
			DataInicioExecucao: resultado.DataInicioExecucao.Format(time.RFC3339),
		})
	}
}

func writeIniciarExecucaoError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "osId")
	case errors.Is(err, domain.ErrMecanicoAutenticadoNaoEncontrado):
		writeProblem(writer, http.StatusForbidden, "Acesso negado", err.Error(), "")
	case errors.Is(err, domain.ErrOSNaoAptaParaExecucao), errors.Is(err, domain.ErrOrcamentoNaoAprovado),
		errors.Is(err, domain.ErrServicosNaoAutorizados), errors.Is(err, domain.ErrRecursosIndisponiveis):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao iniciar execucao", "")
	}
}
