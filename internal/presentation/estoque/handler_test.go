package estoque

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

type registrarEntradaRepositoryStub struct {
	resultado application.Resultado
	err       error
	recebido  application.RegistrarEntradaInput
}

func (stub *registrarEntradaRepositoryStub) RegistrarEntrada(_ context.Context, input application.RegistrarEntradaInput, _ domain.EntradaCadastro) (application.Resultado, error) {
	stub.recebido = input
	return stub.resultado, stub.err
}

const idempotencyKeyValida = "10000000-0000-0000-0000-000000000001"
const itemIDValido = "20000000-0000-0000-0000-000000000001"

func requisicaoValida() string {
	return `{"documentoOrigem":"NF-1","itens":[{"itemId":"` + itemIDValido + `","quantidade":1,"custoUnitario":10}]}`
}

func TestRegistrarEntradaHandlerSucesso(t *testing.T) {
	stub := &registrarEntradaRepositoryStub{resultado: application.Resultado{Entrada: domain.ResultadoEntrada{DocumentoOrigem: "NF-1"}}}
	handler := NewRegistrarEntradaHandler(application.NewRegistrarEntrada(stub))
	request := httptest.NewRequest(http.MethodPost, "/estoque/entradas", strings.NewReader(requisicaoValida()))
	request.Header.Set("Idempotency-Key", idempotencyKeyValida)
	writer := httptest.NewRecorder()
	handler(writer, request)
	if writer.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
	}
	if stub.recebido.IdempotencyKey != idempotencyKeyValida {
		t.Fatalf("idempotencyKey=%q", stub.recebido.IdempotencyKey)
	}
	var resposta domain.ResultadoEntrada
	if err := json.Unmarshal(writer.Body.Bytes(), &resposta); err != nil || resposta.DocumentoOrigem != "NF-1" {
		t.Fatalf("resposta invalida: %s erro=%v", writer.Body.String(), err)
	}
}

func TestRegistrarEntradaHandlerRepeticaoRetorna200(t *testing.T) {
	stub := &registrarEntradaRepositoryStub{resultado: application.Resultado{Entrada: domain.ResultadoEntrada{DocumentoOrigem: "NF-1"}, JaProcessada: true}}
	handler := NewRegistrarEntradaHandler(application.NewRegistrarEntrada(stub))
	request := httptest.NewRequest(http.MethodPost, "/estoque/entradas", strings.NewReader(requisicaoValida()))
	request.Header.Set("Idempotency-Key", idempotencyKeyValida)
	writer := httptest.NewRecorder()
	handler(writer, request)
	if writer.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
	}
}

func TestRegistrarEntradaHandlerSemIdempotencyKey(t *testing.T) {
	handler := NewRegistrarEntradaHandler(application.NewRegistrarEntrada(&registrarEntradaRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/estoque/entradas", strings.NewReader(requisicaoValida()))
	writer := httptest.NewRecorder()
	handler(writer, request)
	if writer.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
	}
}

func TestRegistrarEntradaHandlerCorpoInvalido(t *testing.T) {
	handler := NewRegistrarEntradaHandler(application.NewRegistrarEntrada(&registrarEntradaRepositoryStub{}))
	request := httptest.NewRequest(http.MethodPost, "/estoque/entradas", strings.NewReader("{invalido"))
	request.Header.Set("Idempotency-Key", idempotencyKeyValida)
	writer := httptest.NewRecorder()
	handler(writer, request)
	if writer.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
	}
}

func TestRegistrarEntradaHandlerErroDoRepositorioMapeado(t *testing.T) {
	stub := &registrarEntradaRepositoryStub{err: application.ErrItemInativo}
	handler := NewRegistrarEntradaHandler(application.NewRegistrarEntrada(stub))
	request := httptest.NewRequest(http.MethodPost, "/estoque/entradas", strings.NewReader(requisicaoValida()))
	request.Header.Set("Idempotency-Key", idempotencyKeyValida)
	writer := httptest.NewRecorder()
	handler(writer, request)
	if writer.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
	}
}
