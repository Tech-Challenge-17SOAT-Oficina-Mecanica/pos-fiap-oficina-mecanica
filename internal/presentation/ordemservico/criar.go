package ordemservico

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

type CriarUseCase interface {
	Execute(context.Context, application.CriarInput) (domain.OrdemDeServico, error)
}

type criarRequest struct {
	ClienteID string                    `json:"clienteId"`
	VeiculoID string                    `json:"veiculoId"`
	Servicos  []registrarServicoRequest `json:"servicos"`
	Pecas     []pecaSolicitadaRequest   `json:"pecas"`
}

type pecaSolicitadaRequest struct {
	PecaID     string  `json:"pecaId"`
	Quantidade float64 `json:"quantidade"`
}

func NewCriarHandler(useCase CriarUseCase) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input criarRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		ordem, err := useCase.Execute(request.Context(), application.CriarInput{
			ClienteID: input.ClienteID,
			VeiculoID: input.VeiculoID,
			Servicos:  servicosSolicitados(input.Servicos),
			Pecas:     pecasSolicitadas(input.Pecas),
			UsuarioID: seguranca.UsuarioID(request.Context()),
		})
		if err != nil {
			writeCriarError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(struct {
			OrdemServicoID string `json:"ordemServicoId"`
			ClienteID      string `json:"clienteId"`
			VeiculoID      string `json:"veiculoId"`
			Status         string `json:"status"`
			CriadaEm       string `json:"criadaEm"`
		}{ordem.ID, ordem.ClienteID, ordem.VeiculoID, ordem.Status, ordem.CriadaEm.Format("2006-01-02T15:04:05Z07:00")})
	}
}

func servicosSolicitados(servicos []registrarServicoRequest) []domain.ServicoCadastro {
	cadastros := make([]domain.ServicoCadastro, 0, len(servicos))
	for _, servico := range servicos {
		cadastros = append(cadastros, domain.ServicoCadastro{ServicoID: servico.ServicoID, Observacao: servico.Observacao})
	}
	return cadastros
}

func pecasSolicitadas(pecas []pecaSolicitadaRequest) []application.ItemInput {
	itens := make([]application.ItemInput, 0, len(pecas))
	for _, peca := range pecas {
		itens = append(itens, application.ItemInput{ItemID: peca.PecaID, Quantidade: peca.Quantidade})
	}
	return itens
}

func writeCriarError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrClienteIDObrigatorio), errors.Is(err, application.ErrClienteIDInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "clienteId")
	case errors.Is(err, application.ErrVeiculoIDObrigatorio), errors.Is(err, application.ErrVeiculoIDInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "veiculoId")
	case errors.Is(err, application.ErrServicoIDInvalido), errors.Is(err, domain.ErrServicoDuplicado):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "servicos")
	case errors.Is(err, application.ErrItemIDInvalido):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", "pecaId invalido", "pecas")
	case errors.Is(err, application.ErrItemRepetido), errors.Is(err, domainEstoque.ErrTipoItemInvalido),
		errors.Is(err, domainEstoque.ErrQuantidadeIncompativelComUnidade), strings.Contains(err.Error(), "quantidade"):
		writeProblem(writer, http.StatusBadRequest, "Dados inválidos", err.Error(), "pecas")
	case errors.Is(err, application.ErrClienteNaoEncontrado), errors.Is(err, application.ErrVeiculoNaoEncontrado),
		errors.Is(err, domain.ErrServicoNaoEncontrado), errors.Is(err, application.ErrItemNaoEncontrado):
		writeProblem(writer, http.StatusNotFound, "Recurso não encontrado", err.Error(), "")
	case errors.Is(err, application.ErrVeiculoNaoVinculadoCliente):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "veiculoId")
	case errors.Is(err, domain.ErrServicoInativo), errors.Is(err, application.ErrItemInativo), errors.Is(err, application.ErrItemSemValor):
		writeProblem(writer, http.StatusConflict, "Conflito de estado", err.Error(), "")
	default:
		writeProblem(writer, http.StatusInternalServerError, "Erro interno", "falha ao criar ordem de serviço", "")
	}
}
