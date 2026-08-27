package insumo

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	insumoDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type consultaResponse struct {
	ID                   string       `json:"id"`
	Codigo               string       `json:"codigo"`
	Tipo                 string       `json:"tipo"`
	Nome                 string       `json:"nome"`
	Descricao            string       `json:"descricao"`
	CategoriaID          string       `json:"categoriaId"`
	Categoria            string       `json:"categoria"`
	UnidadeMedida        string       `json:"unidadeMedida"`
	CustoUnitario        *json.Number `json:"custoUnitario"`
	QuantidadeDesejada   *json.Number `json:"quantidadeDesejada,omitempty"`
	SaldoFisico          json.Number  `json:"saldoFisico"`
	SaldoReservado       json.Number  `json:"saldoReservado"`
	SaldoDisponivel      json.Number  `json:"saldoDisponivel"`
	EstoqueMinimo        json.Number  `json:"estoqueMinimo"`
	QuantidadeDisponivel *bool        `json:"quantidadeDisponivel,omitempty"`
	Disponivel           bool         `json:"disponivel"`
	AbaixoDoMinimo       bool         `json:"abaixoDoMinimo"`
	PossuiPedidoEmAberto bool         `json:"possuiPedidoEmAberto"`
	Ativo                bool         `json:"ativo"`
	Version              int          `json:"version"`
}

func NewConsultarInsumosHandler(useCase insumo.ConsultarInsumos) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()

		paginacao, err := sharedhttp.LerPaginacao(query)
		if err != nil {
			problemaDeErro(writer, http.StatusBadRequest, "Dados inválidos", err)
			return
		}

		quantidade, err := lerQuantidade(query.Get("quantidadeDesejada"))
		if err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "quantidadeDesejada")
			return
		}

		somenteDisponiveis, err := sharedhttp.LerBooleano(query, "somenteDisponiveis")
		if err != nil {
			problemaDeErro(writer, http.StatusBadRequest, "Dados inválidos", err)
			return
		}
		incluirInativos, err := sharedhttp.LerBooleano(query, "incluirInativos")
		if err != nil {
			problemaDeErro(writer, http.StatusBadRequest, "Dados inválidos", err)
			return
		}

		resultado, err := useCase.Execute(request.Context(), insumo.FiltrosConsulta{
			Codigo:             query.Get("codigo"),
			Descricao:          query.Get("descricao"),
			CategoriaID:        query.Get("categoriaId"),
			QuantidadeDesejada: quantidade,
			SomenteDisponiveis: somenteDisponiveis,
			IncluirInativos:    incluirInativos,
		}, paginacao.Limit(), paginacao.Offset())
		if err != nil {
			responderErroConsulta(writer, err)
			return
		}

		itens := make([]consultaResponse, 0, len(resultado.Insumos))
		for _, encontrado := range resultado.Insumos {
			itens = append(itens, montarConsultaResponse(encontrado, quantidade))
		}
		sharedhttp.WriteLista(writer, sharedhttp.NovaLista(itens, paginacao, resultado.TotalElementos))
	}
}

func NewConsultarInsumoPorIDHandler(useCase insumo.ConsultarInsumos) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		quantidade, err := lerQuantidade(request.URL.Query().Get("quantidadeDesejada"))
		if err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "quantidadeDesejada")
			return
		}

		encontrado, err := useCase.BuscarPorID(request.Context(), request.PathValue("insumoId"))
		if err != nil {
			responderErroConsulta(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(montarConsultaResponse(encontrado, quantidade))
	}
}

func montarConsultaResponse(encontrado insumoDomain.Insumo, quantidade *string) consultaResponse {
	resposta := consultaResponse{
		ID:                   encontrado.ID,
		Codigo:               encontrado.Codigo,
		Tipo:                 insumoDomain.Tipo,
		Nome:                 encontrado.Nome,
		Descricao:            encontrado.Descricao,
		CategoriaID:          encontrado.CategoriaID,
		Categoria:            encontrado.Categoria,
		UnidadeMedida:        encontrado.UnidadeMedida,
		SaldoFisico:          json.Number(encontrado.SaldoFisico),
		SaldoReservado:       json.Number(encontrado.SaldoReservado),
		SaldoDisponivel:      json.Number(encontrado.SaldoDisponivel()),
		EstoqueMinimo:        json.Number(encontrado.EstoqueMinimo),
		Disponivel:           encontrado.Disponivel(),
		AbaixoDoMinimo:       encontrado.AbaixoDoMinimo(),
		PossuiPedidoEmAberto: encontrado.PossuiPedidoEmAberto,
		Ativo:                encontrado.Ativo,
		Version:              encontrado.Version,
	}
	if encontrado.CustoUnitario != nil {
		custo := json.Number(*encontrado.CustoUnitario)
		resposta.CustoUnitario = &custo
	}
	if quantidade != nil {
		quantidadeJSON := json.Number(*quantidade)
		disponivel := encontrado.AtendeQuantidade(*quantidade)
		resposta.QuantidadeDesejada = &quantidadeJSON
		resposta.QuantidadeDisponivel = &disponivel
	}
	return resposta
}

func lerQuantidade(valor string) (*string, error) {
	if valor == "" {
		return nil, nil
	}
	if !insumo.QuantidadeValida(valor) {
		return nil, insumo.ErrQuantidadeInvalida
	}
	return &valor, nil
}

func responderErroConsulta(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, insumo.ErrInsumoNaoEncontrado):
		problema(writer, http.StatusNotFound, "Não encontrado", err.Error(), "")
	case errors.Is(err, insumo.ErrFiltroObrigatorio):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "descricao")
	case errors.Is(err, insumo.ErrDescricaoCurta):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "descricao")
	case errors.Is(err, insumo.ErrQuantidadeInvalida),
		errors.Is(err, insumo.ErrQuantidadeObrigatoria):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "quantidadeDesejada")
	case errors.Is(err, insumo.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "insumoId")
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao consultar insumos", "")
	}
}
