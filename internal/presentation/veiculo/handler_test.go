package veiculo

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

const veiculoIDTeste = "30000000-0000-0000-0000-000000000001"
const clienteIDTeste = "10000000-0000-0000-0000-000000000001"

type repositoryStub struct {
	veiculo      domain.Veiculo
	err          error
	inativacao   application.Inativacao
	reativacao   application.Reativacao
	ordens       []application.OrdemServicoAberta
	clienteAtivo bool
	placaEmUso   bool
}

func (stub *repositoryStub) CadastrarParaCliente(context.Context, string, domain.Cadastro) (domain.Veiculo, error) {
	return stub.veiculo, stub.err
}

func (stub *repositoryStub) ConsultarPorPlaca(context.Context, string, bool) (domain.Veiculo, error) {
	return stub.veiculo, stub.err
}

func (stub *repositoryStub) Atualizar(context.Context, string, int, domain.Cadastro) (domain.Veiculo, error) {
	return stub.veiculo, stub.err
}

func (stub *repositoryStub) BuscarPorIDIncluindoInativo(context.Context, string) (domain.Veiculo, error) {
	return stub.veiculo, stub.err
}

func (stub *repositoryStub) BuscarOSAbertas(context.Context, string) ([]application.OrdemServicoAberta, error) {
	return stub.ordens, stub.err
}

func (stub *repositoryStub) Inativar(context.Context, application.InativarRepositoryInput) (application.Inativacao, error) {
	return stub.inativacao, stub.err
}

func (stub *repositoryStub) ExisteAtivoPorPlacaExcetoID(context.Context, string, string) (bool, error) {
	return stub.placaEmUso, stub.err
}

func (stub *repositoryStub) ClienteAtivo(context.Context, string) (bool, error) {
	return stub.clienteAtivo, stub.err
}

func (stub *repositoryStub) Reativar(context.Context, string, string) (application.Reativacao, error) {
	return stub.reativacao, stub.err
}

func requisitarVeiculo(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("If-Match", "1")
	writer := httptest.NewRecorder()
	handler.ServeHTTP(writer, request)
	return writer
}

func TestCadastrarVeiculoHandler(t *testing.T) {
	mux := http.NewServeMux()
	stub := &repositoryStub{veiculo: domain.Veiculo{ID: veiculoIDTeste, Cadastro: domain.Cadastro{Placa: "ABC1D23"}}}
	mux.Handle("POST /clientes/{clienteId}/veiculos", NewHandler(application.NewCadastrar(stub)))

	for _, test := range []struct {
		name string
		path string
		body string
		want int
		err  error
	}{
		{"cliente invalido", "/clientes/invalido/veiculos", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusBadRequest, nil},
		{"json invalido", "/clientes/" + clienteIDTeste + "/veiculos", `{`, http.StatusBadRequest, nil},
		{"cadastro invalido", "/clientes/" + clienteIDTeste + "/veiculos", `{"placa":"ABC","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusBadRequest, nil},
		{"cliente ausente", "/clientes/" + clienteIDTeste + "/veiculos", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusNotFound, application.ErrClienteNaoEncontrado},
		{"placa duplicada", "/clientes/" + clienteIDTeste + "/veiculos", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusConflict, application.ErrPlacaDuplicada},
		{"erro interno", "/clientes/" + clienteIDTeste + "/veiculos", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusInternalServerError, errors.New("db")},
		{"sucesso", "/clientes/" + clienteIDTeste + "/veiculos", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusCreated, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub.err = test.err
			writer := requisitarVeiculo(mux, http.MethodPost, test.path, test.body)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
		})
	}
}

func TestAtualizarVeiculoHandler(t *testing.T) {
	mux := http.NewServeMux()
	stub := &repositoryStub{veiculo: domain.Veiculo{ID: veiculoIDTeste, Cadastro: domain.Cadastro{Placa: "ABC1D23"}}}
	mux.Handle("PUT /veiculos/{veiculoId}", NewAtualizarHandler(application.NewAtualizar(stub)))

	t.Run("sem if-match", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/veiculos/"+veiculoIDTeste, bytes.NewBufferString(`{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`))
		writer := httptest.NewRecorder()
		mux.ServeHTTP(writer, request)
		if writer.Code != http.StatusPreconditionRequired {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	for _, test := range []struct {
		name    string
		id      string
		ifMatch string
		body    string
		want    int
		err     error
	}{
		{"id invalido", "invalido", "1", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusBadRequest, nil},
		{"if-match invalido", veiculoIDTeste, "x", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusBadRequest, nil},
		{"json invalido", veiculoIDTeste, "1", `{`, http.StatusBadRequest, nil},
		{"cadastro invalido", veiculoIDTeste, "1", `{"placa":"ABC","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusBadRequest, nil},
		{"nao encontrado", veiculoIDTeste, "1", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusNotFound, application.ErrVeiculoNaoEncontrado},
		{"placa duplicada", veiculoIDTeste, "1", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusConflict, application.ErrPlacaDuplicada},
		{"versao divergente", veiculoIDTeste, "1", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusPreconditionFailed, application.ErrVersaoDivergente},
		{"sucesso", veiculoIDTeste, "1", `{"placa":"ABC1D23","marca":"Fiat","modelo":"Uno","ano":2020}`, http.StatusOK, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub.err = test.err
			request := httptest.NewRequest(http.MethodPut, "/veiculos/"+test.id, strings.NewReader(test.body))
			request.Header.Set("If-Match", test.ifMatch)
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
		})
	}
}

func TestConsultarVeiculoHandler(t *testing.T) {
	mux := http.NewServeMux()
	stub := &repositoryStub{veiculo: domain.Veiculo{ID: veiculoIDTeste, Cadastro: domain.Cadastro{Placa: "ABC1D23"}}}
	mux.Handle("GET /veiculos", NewConsultaHandler(application.NewConsultar(stub)))

	for _, test := range []struct {
		name string
		url  string
		want int
		err  error
	}{
		{"placa invalida", "/veiculos?placa=x&incluirInativos=false", http.StatusBadRequest, nil},
		{"incluir inativos invalido", "/veiculos?placa=ABC1D23&incluirInativos=x", http.StatusBadRequest, nil},
		{"nao encontrado", "/veiculos?placa=ABC1D23&incluirInativos=false", http.StatusNotFound, application.ErrVeiculoNaoEncontrado},
		{"erro interno", "/veiculos?placa=ABC1D23&incluirInativos=false", http.StatusInternalServerError, errors.New("db")},
		{"sucesso", "/veiculos?placa=ABC1D23&incluirInativos=false", http.StatusOK, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub.err = test.err
			writer := requisitarVeiculo(mux, http.MethodGet, test.url, "")
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
		})
	}
}

func TestSituacaoVeiculoHandlers(t *testing.T) {
	inativo := domain.Veiculo{ID: veiculoIDTeste, ClienteID: clienteIDTeste, Cadastro: domain.Cadastro{Placa: "ABC1D23"}, Ativo: false}
	ativo := inativo
	ativo.Ativo = true

	t.Run("inativar sucesso", func(t *testing.T) {
		mux := http.NewServeMux()
		stub := &repositoryStub{veiculo: ativo, inativacao: application.Inativacao{Veiculo: ativo}}
		mux.Handle("DELETE /veiculos/{veiculoId}", NewInativarHandler(application.NewInativar(stub)))
		writer := requisitarVeiculo(mux, http.MethodDelete, "/veiculos/"+veiculoIDTeste, "")
		if writer.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("inativar com os aberta", func(t *testing.T) {
		mux := http.NewServeMux()
		stub := &repositoryStub{veiculo: ativo, ordens: []application.OrdemServicoAberta{{OrdemServicoID: "os-1"}}}
		mux.Handle("DELETE /veiculos/{veiculoId}", NewInativarHandler(application.NewInativar(stub)))
		writer := requisitarVeiculo(mux, http.MethodDelete, "/veiculos/"+veiculoIDTeste, "")
		if writer.Code != http.StatusConflict {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})

	t.Run("reativar sucesso", func(t *testing.T) {
		mux := http.NewServeMux()
		stub := &repositoryStub{
			veiculo: inativo, clienteAtivo: true,
			reativacao: application.Reativacao{Veiculo: inativo, ReativadoEm: time.Now(), ReativadoPor: "usuario"},
		}
		mux.Handle("POST /veiculos/{veiculoId}/reativar", NewReativarHandler(application.NewReativar(stub)))
		writer := requisitarVeiculo(mux, http.MethodPost, "/veiculos/"+veiculoIDTeste+"/reativar", "")
		if writer.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
		}
	})
}
