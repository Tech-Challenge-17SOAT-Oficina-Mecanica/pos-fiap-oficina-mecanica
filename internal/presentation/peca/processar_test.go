package peca

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
)

type processarRepositorioFake struct {
	resultado   application.ResultadoCompraReserva
	err         error
	solicitacao application.SolicitacaoCompraReserva
}

func (fake *processarRepositorioFake) SolicitarCompraEReservar(_ context.Context, solicitacao application.SolicitacaoCompraReserva) (application.ResultadoCompraReserva, error) {
	fake.solicitacao = solicitacao
	return fake.resultado, fake.err
}

func TestSolicitarCompraEReservarPecasHandler(t *testing.T) {
	const corpo = `{"ordemServicoId":"10000000-0000-0000-0000-000000000001","fornecedorId":"20000000-0000-0000-0000-000000000001","itens":[{"itemId":"30000000-0000-0000-0000-000000000001","quantidade":2}]}`

	t.Run("retorna 201", func(t *testing.T) {
		fake := &processarRepositorioFake{resultado: application.ResultadoCompraReserva{
			OrdemServicoID:     "10000000-0000-0000-0000-000000000001",
			StatusOrdemServico: "AGUARDANDO_EXECUCAO",
			PecasReservadas:    []application.ItemReservado{{ItemID: "30000000-0000-0000-0000-000000000001", Quantidade: 2}},
		}}
		resposta := executarProcessamento(fake, corpo, "40000000-0000-0000-0000-000000000001")

		if resposta.Code != http.StatusCreated || !strings.Contains(resposta.Body.String(), `"pecasReservadas"`) {
			t.Fatalf("status=%d body=%s", resposta.Code, resposta.Body.String())
		}
		if fake.solicitacao.HashRequisicao == "" {
			t.Fatal("hash da requisicao nao foi gerado")
		}
	})

	t.Run("reprocessado retorna 200", func(t *testing.T) {
		fake := &processarRepositorioFake{resultado: application.ResultadoCompraReserva{Reprocessado: true}}
		resposta := executarProcessamento(fake, corpo, "40000000-0000-0000-0000-000000000001")
		if resposta.Code != http.StatusOK {
			t.Fatalf("status=%d", resposta.Code)
		}
	})

	t.Run("sem chave retorna 400", func(t *testing.T) {
		resposta := executarProcessamento(&processarRepositorioFake{}, corpo, "")
		if resposta.Code != http.StatusBadRequest || !strings.Contains(resposta.Body.String(), "Idempotency-Key") {
			t.Fatalf("status=%d body=%s", resposta.Code, resposta.Body.String())
		}
	})

	t.Run("json invalido retorna 400", func(t *testing.T) {
		resposta := executarProcessamento(&processarRepositorioFake{}, "{", "40000000-0000-0000-0000-000000000001")
		if resposta.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", resposta.Code)
		}
	})

	t.Run("erro ao ler corpo retorna 400", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/estoque/solicitacoes-compra-reserva", nil)
		request.Header.Set("Idempotency-Key", "40000000-0000-0000-0000-000000000001")
		request.Body = corpoComErro{}
		resposta := httptest.NewRecorder()
		NewSolicitarCompraEReservarPecasHandler(application.NewSolicitarCompraEReservarPecas(&processarRepositorioFake{})).ServeHTTP(resposta, request)
		if resposta.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", resposta.Code)
		}
	})

	t.Run("mapeia erros esperados", func(t *testing.T) {
		casos := []struct {
			nome   string
			err    error
			status int
		}{
			{"fornecedor uuid", application.ErrFornecedorIdentificador, http.StatusBadRequest},
			{"fornecedor ausente", application.ErrFornecedorNaoEncontrado, http.StatusNotFound},
			{"fornecedor inativo", application.ErrFornecedorInativo, http.StatusConflict},
			{"sem item", application.ErrItemObrigatorio, http.StatusBadRequest},
			{"item repetido", application.ErrItemRepetido, http.StatusBadRequest},
			{"quantidade", application.ErrQuantidadeProcessamento, http.StatusBadRequest},
			{"item uuid", application.ErrItemIdentificador, http.StatusBadRequest},
			{"item ausente", application.ErrItemNaoEncontrado, http.StatusNotFound},
			{"item invalido", application.ErrItemProcessamentoInvalido, http.StatusConflict},
			{"os ausente", application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
			{"os invalida", application.ErrOrdemServicoInvalida, http.StatusConflict},
			{"duplicado", application.ErrProcessamentoDuplicado, http.StatusConflict},
			{"chave em uso", application.ErrIdempotencyKeyEmUso, http.StatusConflict},
		}
		for _, caso := range casos {
			t.Run(caso.nome, func(t *testing.T) {
				fake := &processarRepositorioFake{err: caso.err}
				resposta := executarProcessamento(fake, corpo, "40000000-0000-0000-0000-000000000001")
				if resposta.Code != caso.status {
					t.Fatalf("status=%d body=%s", resposta.Code, resposta.Body.String())
				}
			})
		}
	})

	t.Run("erro inesperado retorna 500", func(t *testing.T) {
		fake := &processarRepositorioFake{err: errors.New("falha")}
		resposta := executarProcessamento(fake, corpo, "40000000-0000-0000-0000-000000000001")
		if resposta.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d", resposta.Code)
		}
	})
}

func executarProcessamento(fake *processarRepositorioFake, corpo, chave string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/estoque/solicitacoes-compra-reserva", strings.NewReader(corpo))
	if chave != "" {
		request.Header.Set("Idempotency-Key", chave)
	}
	response := httptest.NewRecorder()
	NewSolicitarCompraEReservarPecasHandler(application.NewSolicitarCompraEReservarPecas(fake)).ServeHTTP(response, request)
	return response
}

type corpoComErro struct{}

func (corpoComErro) Read([]byte) (int, error) { return 0, errors.New("falha") }
func (corpoComErro) Close() error             { return nil }

var _ io.ReadCloser = corpoComErro{}
