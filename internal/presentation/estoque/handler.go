package estoque

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type itemEntradaRequest struct {
	ItemID        string  `json:"itemId"`
	Quantidade    float64 `json:"quantidade"`
	CustoUnitario float64 `json:"custoUnitario"`
}

type registrarEntradaRequest struct {
	DocumentoOrigem      string               `json:"documentoOrigem"`
	FornecedorID         string               `json:"fornecedorId"`
	PedidoCompraID       string               `json:"pedidoCompraId"`
	ConfirmarDivergencia bool                 `json:"confirmarDivergencia"`
	Itens                []itemEntradaRequest `json:"itens"`
}

type itemSaidaRequest struct {
	ItemID     string  `json:"itemId"`
	Quantidade float64 `json:"quantidade"`
}

type registrarSaidaRequest struct {
	OrdemServicoID       string             `json:"ordemServicoId"`
	Itens                []itemSaidaRequest `json:"itens"`
	LiberarSaldoNaoUsado *bool              `json:"liberarSaldoNaoUsado"`
}

func NewRegistrarEntradaHandler(useCase application.RegistrarEntrada) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		idempotencyKey := request.Header.Get("Idempotency-Key")

		var body registrarEntradaRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}

		itens := make([]application.ItemInput, 0, len(body.Itens))
		for _, item := range body.Itens {
			itens = append(itens, application.ItemInput{ItemID: item.ItemID, Quantidade: item.Quantidade, CustoUnitario: item.CustoUnitario})
		}

		resultado, err := useCase.Execute(request.Context(), application.RegistrarEntradaInput{
			IdempotencyKey: idempotencyKey, DocumentoOrigem: body.DocumentoOrigem, FornecedorID: body.FornecedorID,
			PedidoCompraID: body.PedidoCompraID, ConfirmarDivergencia: body.ConfirmarDivergencia, Itens: itens,
			UsuarioID: seguranca.UsuarioID(request.Context()),
		})
		if err != nil {
			writeEntradaError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if resultado.JaProcessada {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusCreated)
		}
		_ = json.NewEncoder(writer).Encode(resultado.Entrada)
	}
}

func NewRegistrarSaidaHandler(useCase application.RegistrarSaida) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		idempotencyKey := request.Header.Get("Idempotency-Key")

		var body registrarSaidaRequest
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(writer, http.StatusBadRequest, "Dados invalidos", "corpo da requisicao invalido", "")
			return
		}

		itens := make([]application.ItemSaidaInput, 0, len(body.Itens))
		for _, item := range body.Itens {
			itens = append(itens, application.ItemSaidaInput{ItemID: item.ItemID, Quantidade: item.Quantidade})
		}
		liberarSaldoNaoUsado := true
		if body.LiberarSaldoNaoUsado != nil {
			liberarSaldoNaoUsado = *body.LiberarSaldoNaoUsado
		}

		resultado, err := useCase.Execute(request.Context(), application.RegistrarSaidaInput{
			IdempotencyKey: idempotencyKey, OrdemServicoID: body.OrdemServicoID, Itens: itens,
			LiberarSaldoNaoUsado: liberarSaldoNaoUsado, UsuarioID: seguranca.UsuarioID(request.Context()),
		})
		if err != nil {
			writeSaidaError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if resultado.JaProcessada {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusCreated)
		}
		_ = json.NewEncoder(writer).Encode(resultado.Saida)
	}
}

func writeEntradaError(writer http.ResponseWriter, err error) {
	status, title, campo := http.StatusInternalServerError, "Erro interno", ""
	switch {
	case errors.Is(err, domain.ErrIdempotencyKeyObrigatoria):
		status, title, campo = http.StatusBadRequest, "Dados invalidos", "Idempotency-Key"
	case errors.Is(err, domain.ErrFornecedorIDInvalido):
		status, title, campo = http.StatusBadRequest, "Dados invalidos", "fornecedorId"
	case errors.Is(err, domain.ErrDocumentoOrigemObrigatorio), errors.Is(err, domain.ErrItensObrigatorios),
		errors.Is(err, domain.ErrItensExcedemLimite), errors.Is(err, domain.ErrItemIDInvalido),
		errors.Is(err, domain.ErrItemRepetido), errors.Is(err, domain.ErrCustoInvalido),
		errors.Is(err, domain.ErrQuantidadeIncompativelComUnidade):
		status, title = http.StatusBadRequest, "Dados invalidos"
	case errors.Is(err, application.ErrItemNaoEncontrado), errors.Is(err, application.ErrPedidoCompraNaoEncontrado),
		errors.Is(err, application.ErrFornecedorNaoEncontrado):
		status, title = http.StatusNotFound, "Recurso nao encontrado"
	case errors.Is(err, application.ErrItemInativo), errors.Is(err, application.ErrFornecedorInativo),
		errors.Is(err, application.ErrDocumentoOrigemDuplicado), errors.Is(err, application.ErrFornecedorDivergente),
		errors.Is(err, application.ErrItemForaDoPedido), errors.Is(err, application.ErrDivergenciaQuantidade):
		status, title = http.StatusConflict, "Conflito de estado"
	default:
		if err.Error() == "quantidade deve ser maior que zero" || err.Error() == "a quantidade de peca deve ser inteira" {
			status, title = http.StatusBadRequest, "Dados invalidos"
		}
	}
	writeProblem(writer, status, title, err.Error(), campo)
}

func writeSaidaError(writer http.ResponseWriter, err error) {
	status, title, campo := http.StatusInternalServerError, "Erro interno", ""
	switch {
	case errors.Is(err, domain.ErrIdempotencyKeyObrigatoria):
		status, title, campo = http.StatusBadRequest, "Dados invalidos", "Idempotency-Key"
	case errors.Is(err, domain.ErrItensObrigatorios), errors.Is(err, domain.ErrItensExcedemLimite),
		errors.Is(err, domain.ErrItemIDInvalido), errors.Is(err, domain.ErrItemRepetido),
		errors.Is(err, domain.ErrQuantidadeIncompativelComUnidade):
		status, title = http.StatusBadRequest, "Dados invalidos"
	case errors.Is(err, application.ErrOrdemServicoNaoEncontrada), errors.Is(err, application.ErrItemNaoEncontrado):
		status, title = http.StatusNotFound, "Recurso nao encontrado"
	case errors.Is(err, application.ErrOSForaDeExecucao), errors.Is(err, application.ErrReservaAtivaNaoEncontrada),
		errors.Is(err, application.ErrSaldoInsuficiente), errors.Is(err, domain.ErrQuantidadeMaiorQueReserva):
		status, title = http.StatusConflict, "Conflito de estado"
	default:
		if err.Error() == "quantidade deve ser maior que zero" || err.Error() == "a quantidade de peca deve ser inteira" {
			status, title = http.StatusBadRequest, "Dados invalidos"
		}
	}
	writeProblem(writer, status, title, err.Error(), campo)
}

func writeProblem(writer http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/estoque", Title: title, Status: status, Detail: detail}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(writer, problem)
}
