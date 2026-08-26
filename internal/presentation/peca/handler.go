package peca

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type pecaResponse struct {
	ID                   string       `json:"id"`
	Codigo               string       `json:"codigo"`
	Tipo                 string       `json:"tipo"`
	Nome                 string       `json:"nome"`
	Descricao            string       `json:"descricao"`
	CategoriaID          string       `json:"categoriaId"`
	Categoria            string       `json:"categoria"`
	Fabricante           *string      `json:"fabricante"`
	UnidadeMedida        string       `json:"unidadeMedida"`
	PrecoVenda           *json.Number `json:"precoVenda"`
	SaldoFisico          int64        `json:"saldoFisico"`
	SaldoReservado       int64        `json:"saldoReservado"`
	SaldoDisponivel      int64        `json:"saldoDisponivel"`
	QuantidadeDesejada   *int64       `json:"quantidadeDesejada,omitempty"`
	QuantidadeDisponivel *bool        `json:"quantidadeDisponivel,omitempty"`
	EstoqueMinimo        int64        `json:"estoqueMinimo"`
	Disponivel           bool         `json:"disponivel"`
	AbaixoDoMinimo       bool         `json:"abaixoDoMinimo"`
	PossuiPedidoEmAberto bool         `json:"possuiPedidoEmAberto"`
	Ativo                bool         `json:"ativo"`
	Version              int          `json:"version"`
	// Só o cadastro devolve dataCriacao; consultar e deletar a omitem.
	DataCriacao *time.Time `json:"dataCriacao,omitempty"`
}

func NewConsultarPecasHandler(useCase peca.ConsultarPecas) http.HandlerFunc {
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

		filtros := peca.Filtros{
			Codigo:             query.Get("codigo"),
			Descricao:          query.Get("descricao"),
			CategoriaID:        query.Get("categoriaId"),
			Fabricante:         query.Get("fabricante"),
			SomenteDisponiveis: somenteDisponiveis,
			IncluirInativos:    incluirInativos,
			QuantidadeDesejada: quantidade,
		}

		resultado, err := useCase.Execute(request.Context(), filtros, paginacao.Limit(), paginacao.Offset())
		if err != nil {
			responderErro(writer, err)
			return
		}

		itens := make([]pecaResponse, 0, len(resultado.Pecas))
		for _, encontrada := range resultado.Pecas {
			itens = append(itens, montarResponse(encontrada, quantidade))
		}
		sharedhttp.WriteLista(writer, sharedhttp.NovaLista(itens, paginacao, resultado.TotalElementos))
	}
}

func NewConsultarPecaPorIDHandler(useCase peca.ConsultarPecas) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		quantidade, err := lerQuantidade(request.URL.Query().Get("quantidadeDesejada"))
		if err != nil {
			problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "quantidadeDesejada")
			return
		}

		encontrada, err := useCase.BuscarPorID(request.Context(), request.PathValue("pecaId"))
		if err != nil {
			responderErro(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(montarResponse(encontrada, quantidade))
	}
}

func montarResponse(encontrada pecaDomain.Peca, quantidade *int64) pecaResponse {
	resposta := pecaResponse{
		ID:                   encontrada.ID,
		Codigo:               encontrada.Codigo,
		Tipo:                 "PECA",
		Nome:                 encontrada.Nome,
		Descricao:            encontrada.Descricao,
		CategoriaID:          encontrada.CategoriaID,
		Categoria:            encontrada.Categoria,
		Fabricante:           encontrada.Fabricante,
		UnidadeMedida:        encontrada.UnidadeMedida,
		SaldoFisico:          encontrada.SaldoFisico,
		SaldoReservado:       encontrada.SaldoReservado,
		SaldoDisponivel:      encontrada.SaldoDisponivel(),
		EstoqueMinimo:        encontrada.EstoqueMinimo,
		Disponivel:           encontrada.Disponivel(),
		AbaixoDoMinimo:       encontrada.AbaixoDoMinimo(),
		PossuiPedidoEmAberto: encontrada.PossuiPedidoEmAberto,
		Ativo:                encontrada.Ativo,
		Version:              encontrada.Version,
		DataCriacao:          encontrada.DataCriacao,
	}
	if encontrada.PrecoVenda != nil {
		preco := json.Number(*encontrada.PrecoVenda)
		resposta.PrecoVenda = &preco
	}
	if quantidade != nil {
		atende := encontrada.AtendeQuantidade(*quantidade)
		resposta.QuantidadeDesejada = quantidade
		resposta.QuantidadeDisponivel = &atende
	}
	return resposta
}

func lerQuantidade(valor string) (*int64, error) {
	if valor == "" {
		return nil, nil
	}
	quantidade, err := strconv.ParseInt(valor, 10, 64)
	if err != nil {
		return nil, peca.ErrQuantidadeInvalida
	}
	return &quantidade, nil
}

func responderErro(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, peca.ErrNaoEncontrada):
		problema(writer, http.StatusNotFound, "Não encontrado", err.Error(), "")
	case errors.Is(err, peca.ErrFiltroObrigatorio):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "descricao")
	case errors.Is(err, peca.ErrDescricaoCurta):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "descricao")
	case errors.Is(err, peca.ErrQuantidadeInvalida):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "quantidadeDesejada")
	case errors.Is(err, peca.ErrIdentificadorInvalido):
		problema(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "pecaId")
	default:
		problema(writer, http.StatusInternalServerError, "Erro interno", "falha ao consultar peças", "")
	}
}

func problema(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{
		Type:   "https://api.oficina-mecanica.dev/errors/estoque",
		Title:  title,
		Status: status,
		Detail: detail,
	}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}

// problemaDeErro usa o campo que o proprio erro carrega, quando ele carrega.
func problemaDeErro(writer http.ResponseWriter, status int, title string, err error) {
	problema(writer, status, title, err.Error(), sharedhttp.CampoDoErro(err))
}
