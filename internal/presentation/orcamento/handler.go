package orcamento

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

type aprovacaoResponse struct {
	OrcamentoID         string `json:"orcamentoId"`
	OrdemServicoID      string `json:"ordemServicoId"`
	TipoOrcamento       string `json:"tipoOrcamento"`
	OrcamentoOriginalID string `json:"orcamentoOriginalId,omitempty"`
	StatusOrcamento     string `json:"statusOrcamento"`
	StatusOrdemServico  string `json:"statusOrdemServico"`
	ClienteID           string `json:"clienteId"`
	DataAprovacao       string `json:"dataAprovacao"`
}

func NewAprovarHandler(useCase application.Aprovar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		claims := seguranca.Claims(request.Context())
		usuarioID := ""
		if claims.ClienteID == "" {
			usuarioID = claims.UsuarioID
		}
		resultado, err := useCase.Execute(request.Context(), application.AprovarInput{
			OrcamentoID:    request.PathValue("orcamentoId"),
			ClienteID:      claims.ClienteID,
			UsuarioID:      usuarioID,
			OrdemServicoID: claims.OrdemServicoID,
		})
		if err != nil {
			responderErroAprovacao(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(aprovacaoResponse{
			OrcamentoID:         resultado.OrcamentoID,
			OrdemServicoID:      resultado.OrdemServicoID,
			TipoOrcamento:       resultado.TipoOrcamento,
			OrcamentoOriginalID: resultado.OrcamentoOriginalID,
			StatusOrcamento:     resultado.StatusOrcamento,
			StatusOrdemServico:  resultado.StatusOrdemServico,
			ClienteID:           resultado.ClienteID,
			DataAprovacao:       resultado.DataAprovacao.Format(timeFormat),
		})
	}
}

func responderErroAprovacao(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrIdentificadorInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "orcamentoId")
	case errors.Is(err, application.ErrOrcamentoNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "orcamentoId")
	case errors.Is(err, application.ErrAcessoNegado):
		writeProblem(writer, http.StatusForbidden, "Acesso negado", err.Error(), "")
	case errors.Is(err, application.ErrOrcamentoJaDecidido),
		errors.Is(err, application.ErrOrdemServicoStatusInvalido),
		errors.Is(err, application.ErrOrcamentoComplementarSemPai):
		writeProblem(writer, http.StatusConflict, "Conflito", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao aprovar orcamento", "")
	}
}

func NewConsultarHandler(useCase application.Consultar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ordemServicoID := request.PathValue("osId")
		if !validation.IsUUID(ordemServicoID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "osId invalido", "osId")
			return
		}
		claims := seguranca.Claims(request.Context())
		if claims.ClienteID != "" && claims.OrdemServicoID != ordemServicoID {
			writeProblem(writer, http.StatusForbidden, "Acesso negado", "cliente sem acesso ao orcamento desta ordem de servico", "")
			return
		}
		consulta, err := useCase.Execute(request.Context(), ordemServicoID, claims.ClienteID)
		if err != nil {
			if errors.Is(err, application.ErrOrdemServicoNaoEncontrada) || errors.Is(err, application.ErrOrcamentoNaoEncontrado) {
				writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "osId")
				return
			}
			if errors.Is(err, application.ErrAcessoNegado) {
				writeProblem(writer, http.StatusForbidden, "Acesso negado", err.Error(), "")
				return
			}
			writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao consultar orcamento", "")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(newResponse(consulta))
	}
}

type response struct {
	Cliente               clienteResponse     `json:"cliente"`
	OrdemServicoID        string              `json:"ordemServicoId"`
	StatusOrdemServico    string              `json:"statusOrdemServico"`
	ValorTotalGeral       float64             `json:"valorTotalGeral"`
	EstimativaEntregaDias *int                `json:"estimativaEntregaDias,omitempty"`
	Orcamentos            []orcamentoResponse `json:"orcamentos"`
}

type clienteResponse struct {
	ClienteID     string `json:"clienteId"`
	Nome          string `json:"nome"`
	Documento     string `json:"documento"`
	TipoDocumento string `json:"tipoDocumento"`
}

type orcamentoResponse struct {
	OrcamentoID           string             `json:"orcamentoId"`
	OrcamentoOriginalID   string             `json:"orcamentoOriginalId,omitempty"`
	TipoOrcamento         string             `json:"tipoOrcamento"`
	StatusOrcamento       string             `json:"statusOrcamento"`
	EstimativaEntregaDias *int               `json:"estimativaEntregaDias,omitempty"`
	DataGeracao           string             `json:"dataGeracao"`
	Itens                 []itemResponse     `json:"itens"`
	Problemas             []problemaResponse `json:"problemas"`
	ValorTotal            float64            `json:"valorTotal"`
}

type problemaResponse struct {
	ProblemaID   string `json:"problemaId"`
	Descricao    string `json:"descricao"`
	Observacoes  string `json:"observacoes,omitempty"`
	RegistradoEm string `json:"registradoEm"`
}

type itemResponse struct {
	Tipo          string  `json:"tipo"`
	Descricao     string  `json:"descricao"`
	Quantidade    float64 `json:"quantidade"`
	ValorUnitario float64 `json:"valorUnitario"`
	ValorTotal    float64 `json:"valorTotal"`
}

func newResponse(consulta domain.Consulta) response {
	result := response{Cliente: clienteResponse{ClienteID: consulta.Cliente.ID, Nome: consulta.Cliente.Nome, Documento: validation.MascararDocumento(consulta.Cliente.Documento, consulta.Cliente.TipoDocumento), TipoDocumento: consulta.Cliente.TipoDocumento}, OrdemServicoID: consulta.OrdemServicoID, StatusOrdemServico: consulta.StatusOrdemServico, ValorTotalGeral: consulta.ValorTotalGeral, EstimativaEntregaDias: consulta.EstimativaEntregaDias, Orcamentos: make([]orcamentoResponse, 0, len(consulta.Orcamentos))}
	for _, budget := range consulta.Orcamentos {
		out := orcamentoResponse{OrcamentoID: budget.ID, OrcamentoOriginalID: budget.OriginalID, TipoOrcamento: budget.Tipo, StatusOrcamento: budget.Status, EstimativaEntregaDias: budget.EstimativaDias, DataGeracao: budget.DataGeracao.Format(timeFormat), Itens: make([]itemResponse, 0, len(budget.Itens)), Problemas: make([]problemaResponse, 0, len(budget.Problemas)), ValorTotal: budget.ValorTotal}
		for _, item := range budget.Itens {
			out.Itens = append(out.Itens, itemResponse{Tipo: item.Tipo, Descricao: item.Descricao, Quantidade: item.Quantidade, ValorUnitario: item.ValorUnitario, ValorTotal: item.ValorTotal})
		}
		for _, problema := range budget.Problemas {
			out.Problemas = append(out.Problemas, problemaResponse{ProblemaID: problema.ID, Descricao: problema.Descricao, Observacoes: problema.Observacoes, RegistradoEm: problema.RegistradoEm.Format(timeFormat)})
		}
		result.Orcamentos = append(result.Orcamentos, out)
	}
	return result
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

type recusarRequest struct {
	Motivo string `json:"motivo"`
}

type recusarResponse struct {
	OrcamentoID         string `json:"orcamentoId"`
	OrdemServicoID      string `json:"ordemServicoId"`
	TipoOrcamento       string `json:"tipoOrcamento"`
	OrcamentoOriginalID string `json:"orcamentoOriginalId,omitempty"`
	StatusOrcamento     string `json:"statusOrcamento"`
	StatusOrdemServico  string `json:"statusOrdemServico"`
	ClienteID           string `json:"clienteId"`
	DataRecusa          string `json:"dataRecusa"`
	Motivo              string `json:"motivo,omitempty"`
}

func NewRecusarHandler(useCase application.Recusar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		orcamentoID := request.PathValue("orcamentoId")
		if !validation.IsUUID(orcamentoID) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "orcamentoId invalido", "orcamentoId")
			return
		}
		var input recusarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil && !errors.Is(err, io.EOF) {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}
		claims := seguranca.Claims(request.Context())
		decisao, err := useCase.Execute(request.Context(), application.RecusarInput{
			OrcamentoID:    orcamentoID,
			ClienteID:      claims.ClienteID,
			OrdemServicoID: claims.OrdemServicoID,
			Motivo:         input.Motivo,
		})
		if err != nil {
			writeRecusarError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(recusarResponse{
			OrcamentoID:         decisao.OrcamentoID,
			OrdemServicoID:      decisao.OrdemServicoID,
			TipoOrcamento:       decisao.TipoOrcamento,
			OrcamentoOriginalID: decisao.OrcamentoOriginalID,
			StatusOrcamento:     decisao.StatusOrcamento,
			StatusOrdemServico:  decisao.StatusOrdemServico,
			ClienteID:           decisao.ClienteID,
			DataRecusa:          decisao.DecididoEm.Format(timeFormat),
			Motivo:              decisao.Motivo,
		})
	}
}

func writeRecusarError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrOrcamentoNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso nao encontrado", err.Error(), "orcamentoId")
	case errors.Is(err, application.ErrAcessoNegado):
		writeProblem(writer, http.StatusForbidden, "Acesso negado", err.Error(), "")
	case errors.Is(err, application.ErrOrcamentoJaDecidido), errors.Is(err, application.ErrOrcamentoComplementarSemPrincipal), errors.Is(err, application.ErrOrdemServicoNaoAguardandoAprovacao):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	case errors.Is(err, application.ErrMotivoInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados invalidos", err.Error(), "motivo")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "erro ao recusar orcamento", "")
	}
}

func writeProblem(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/orcamento", Title: title, Status: status, Detail: detail}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}
