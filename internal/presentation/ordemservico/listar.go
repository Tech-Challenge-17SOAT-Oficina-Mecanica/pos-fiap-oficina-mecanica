package ordemservico

import (
	"errors"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type listagemClienteResponse struct {
	ClienteID string `json:"clienteId"`
	Nome      string `json:"nome"`
	Documento string `json:"documento"`
}

type listagemVeiculoResponse struct {
	VeiculoID string `json:"veiculoId"`
	Placa     string `json:"placa"`
	Marca     string `json:"marca"`
	Modelo    string `json:"modelo"`
}

type listagemItemResponse struct {
	OrdemServicoID string                  `json:"ordemServicoId"`
	Cliente        listagemClienteResponse `json:"cliente"`
	Veiculo        listagemVeiculoResponse `json:"veiculo"`
	Status         string                  `json:"status"`
}

func NewListarHandler(useCase application.Listar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		paginacao, err := sharedhttp.LerPaginacao(query)
		if err != nil {
			writeCampoInvalido(writer, err)
			return
		}
		status := query.Get("status")
		if status != "" && !domain.StatusListagemValido(status) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "status invalido", "status")
			return
		}

		resultado, err := useCase.Execute(request.Context(), application.ListarInput{
			Status: status, Documento: query.Get("documento"), Placa: query.Get("placa"),
			Limite: paginacao.Limit(), Deslocamento: paginacao.Offset(),
		})
		if err != nil {
			writeListarError(writer, err)
			return
		}

		itens := make([]listagemItemResponse, 0, len(resultado.Itens))
		for _, item := range resultado.Itens {
			itens = append(itens, listagemItemResponse{
				OrdemServicoID: item.OrdemServicoID,
				Cliente:        listagemClienteResponse{ClienteID: item.ClienteID, Nome: item.ClienteNome, Documento: item.ClienteDocumento},
				Veiculo:        listagemVeiculoResponse{VeiculoID: item.VeiculoID, Placa: item.Placa, Marca: item.Marca, Modelo: item.Modelo},
				Status:         item.Status,
			})
		}
		sharedhttp.WriteLista(writer, sharedhttp.NovaLista(itens, paginacao, resultado.TotalElementos))
	}
}

func writeCampoInvalido(writer http.ResponseWriter, err error) {
	writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), sharedhttp.CampoDoErro(err))
}

func writeListarError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrStatusInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "status")
	case errors.Is(err, application.ErrDocumentoInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "documento")
	case errors.Is(err, application.ErrPlacaInvalida):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "placa")
	case errors.Is(err, application.ErrClienteNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "documento")
	case errors.Is(err, application.ErrVeiculoNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "placa")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao listar ordens de servico", "")
	}
}
