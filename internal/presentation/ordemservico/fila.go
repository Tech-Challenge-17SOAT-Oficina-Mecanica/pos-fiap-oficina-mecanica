package ordemservico

import (
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type filaVeiculoResponse struct {
	Placa  string `json:"placa"`
	Marca  string `json:"marca"`
	Modelo string `json:"modelo"`
}

type filaItemResponse struct {
	OrdemServicoID        string              `json:"ordemServicoId"`
	Veiculo               filaVeiculoResponse `json:"veiculo"`
	Status                string              `json:"status"`
	MecanicoResponsavelID *string             `json:"mecanicoResponsavelId"`
	DataEntradaFila       time.Time           `json:"dataEntradaFila"`
}

func NewConsultarFilaHandler(useCase application.ConsultarFila) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		paginacao, err := sharedhttp.LerPaginacao(request.URL.Query())
		if err != nil {
			writeCampoInvalido(writer, err)
			return
		}
		resultado, err := useCase.Execute(request.Context(), application.ConsultarFilaInput{
			Limite: paginacao.Limit(), Deslocamento: paginacao.Offset(),
		})
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao consultar fila de atendimento", "")
			return
		}
		itens := make([]filaItemResponse, 0, len(resultado.Itens))
		for _, item := range resultado.Itens {
			itens = append(itens, filaItemResponse{
				OrdemServicoID: item.OrdemServicoID,
				Veiculo:        filaVeiculoResponse{Placa: item.Placa, Marca: item.Marca, Modelo: item.Modelo},
				Status:         item.Status, MecanicoResponsavelID: item.MecanicoResponsavelID,
				DataEntradaFila: item.DataEntradaFila,
			})
		}
		sharedhttp.WriteLista(writer, sharedhttp.NovaLista(itens, paginacao, resultado.TotalElementos))
	}
}
