package ordemservico

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type clienteResumoResponse struct {
	ClienteID string `json:"clienteId"`
	Nome      string `json:"nome"`
	Documento string `json:"documento"`
}

type veiculoResumoResponse struct {
	VeiculoID string `json:"veiculoId"`
	Placa     string `json:"placa"`
	Marca     string `json:"marca"`
	Modelo    string `json:"modelo"`
	Ano       int    `json:"ano"`
}

type problemaConsultaResponse struct {
	ProblemaID     string `json:"problemaId"`
	Descricao      string `json:"descricao"`
	OrcamentoID    string `json:"orcamentoId,omitempty"`
	IdentificadoEm string `json:"identificadoEm"`
}

type itemOrcamentoConsultaResponse struct {
	Tipo          string  `json:"tipo"`
	Descricao     string  `json:"descricao"`
	Quantidade    float64 `json:"quantidade"`
	ValorUnitario float64 `json:"valorUnitario"`
	ValorTotal    float64 `json:"valorTotal"`
}

type orcamentoConsultaResponse struct {
	OrcamentoID         string                          `json:"orcamentoId"`
	Tipo                string                          `json:"tipo"`
	OrcamentoOriginalID string                          `json:"orcamentoOriginalId,omitempty"`
	Itens               []itemOrcamentoConsultaResponse `json:"itens"`
	ValorTotal          float64                         `json:"valorTotal"`
	DataGeracao         string                          `json:"dataGeracao"`
}

type eventoConsultaResponse struct {
	ID             string `json:"id"`
	Agregado       string `json:"agregado"`
	AgregadoID     string `json:"agregadoId"`
	OrdemServico   string `json:"ordemServicoId"`
	TipoEvento     string `json:"tipoEvento"`
	StatusAnterior string `json:"statusAnterior,omitempty"`
	StatusNovo     string `json:"statusNovo,omitempty"`
	OcorridoEm     string `json:"ocorridoEm"`
	RegistradoEm   string `json:"registradoEm"`
}

type consultaResponse struct {
	OrdemServicoID     string                      `json:"ordemServicoId"`
	StatusOrdemServico string                      `json:"statusOrdemServico"`
	Cliente            clienteResumoResponse       `json:"cliente"`
	Veiculo            veiculoResumoResponse       `json:"veiculo"`
	Problemas          []problemaConsultaResponse  `json:"problemas"`
	Orcamentos         []orcamentoConsultaResponse `json:"orcamentos"`
	ValorTotalGeral    float64                     `json:"valorTotalGeral"`
	Eventos            []eventoConsultaResponse    `json:"eventos"`
}

func NewConsultarHandler(useCase application.Consultar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		osID := request.PathValue("osId")
		if !validation.IsUUID(osID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		claims := seguranca.Claims(request.Context())
		if claims.ClienteID != "" && claims.OrdemServicoID != osID {
			writeProblem(writer, http.StatusForbidden, "Acesso negado", "cliente sem acesso a esta ordem de servico", "")
			return
		}
		consulta, err := useCase.Execute(request.Context(), osID, claims.ClienteID)
		if err != nil {
			if errors.Is(err, application.ErrOrdemServicoNaoEncontrada) {
				writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "osId")
				return
			}
			if errors.Is(err, application.ErrAcessoNegado) {
				writeProblem(writer, http.StatusForbidden, "Acesso negado", err.Error(), "")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao consultar ordem de servico", "")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(newConsultaResponse(consulta))
	}
}

func newConsultaResponse(consulta domain.ConsultaDetalhada) consultaResponse {
	response := consultaResponse{
		OrdemServicoID:     consulta.OrdemServicoID,
		StatusOrdemServico: consulta.StatusOrdemServico,
		Cliente:            clienteResumoResponse{ClienteID: consulta.Cliente.ID, Nome: consulta.Cliente.Nome, Documento: mascararDocumento(consulta.Cliente.Documento)},
		Veiculo:            veiculoResumoResponse{VeiculoID: consulta.Veiculo.ID, Placa: consulta.Veiculo.Placa, Marca: consulta.Veiculo.Marca, Modelo: consulta.Veiculo.Modelo, Ano: consulta.Veiculo.Ano},
		Problemas:          make([]problemaConsultaResponse, 0, len(consulta.Problemas)),
		Orcamentos:         make([]orcamentoConsultaResponse, 0, len(consulta.Orcamentos)),
		ValorTotalGeral:    consulta.ValorTotalGeral,
		Eventos:            make([]eventoConsultaResponse, 0, len(consulta.Eventos)),
	}
	for _, problema := range consulta.Problemas {
		response.Problemas = append(response.Problemas, problemaConsultaResponse{
			ProblemaID: problema.ID, Descricao: problema.Descricao, OrcamentoID: problema.OrcamentoID,
			IdentificadoEm: problema.IdentificadoEm.Format(time.RFC3339),
		})
	}
	for _, orcamento := range consulta.Orcamentos {
		itens := make([]itemOrcamentoConsultaResponse, 0, len(orcamento.Itens))
		for _, item := range orcamento.Itens {
			itens = append(itens, itemOrcamentoConsultaResponse{
				Tipo: item.Tipo, Descricao: item.Descricao, Quantidade: item.Quantidade,
				ValorUnitario: item.ValorUnitario, ValorTotal: item.ValorTotal,
			})
		}
		response.Orcamentos = append(response.Orcamentos, orcamentoConsultaResponse{
			OrcamentoID: orcamento.ID, Tipo: orcamento.Tipo, OrcamentoOriginalID: orcamento.OrcamentoOriginalID,
			Itens: itens, ValorTotal: orcamento.ValorTotal, DataGeracao: orcamento.DataGeracao.Format(time.RFC3339),
		})
	}
	for _, evento := range consulta.Eventos {
		statusAnterior, statusNovo := extrairStatusTransicao(evento.Dados)
		response.Eventos = append(response.Eventos, eventoConsultaResponse{
			ID: evento.ID, Agregado: evento.Agregado, AgregadoID: evento.AgregadoID, OrdemServico: consulta.OrdemServicoID,
			TipoEvento: evento.TipoEvento, StatusAnterior: statusAnterior, StatusNovo: statusNovo,
			OcorridoEm: evento.OcorridoEm.Format(time.RFC3339), RegistradoEm: evento.RegistradoEm.Format(time.RFC3339),
		})
	}
	return response
}

func mascararDocumento(documento string) string {
	digitos := validation.OnlyDigits(documento)
	switch len(digitos) {
	case 11:
		return validation.MascararDocumento(documento, "CPF")
	case 14:
		return validation.MascararDocumento(documento, "CNPJ")
	default:
		return documento
	}
}

func extrairStatusTransicao(payload json.RawMessage) (string, string) {
	if len(payload) == 0 {
		return "", ""
	}
	var dados map[string]any
	if err := json.Unmarshal(payload, &dados); err != nil {
		return "", ""
	}
	statusAnterior, _ := dados["statusAnterior"].(string)
	statusNovo, _ := dados["statusNovo"].(string)
	return statusAnterior, statusNovo
}
