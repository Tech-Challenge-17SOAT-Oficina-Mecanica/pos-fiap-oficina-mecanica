package orcamento

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	notificacaoDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
)

type repositoryStub struct {
	result domain.Consulta
	err    error
}

func (stub repositoryStub) Consultar(context.Context, string, string) (domain.Consulta, error) {
	return stub.result, stub.err
}

type aprovarStub struct {
	result domain.Aprovacao
	err    error
	input  application.AprovarInput
}

func (stub *aprovarStub) Aprovar(_ context.Context, input application.AprovarInput) (domain.Aprovacao, error) {
	stub.input = input
	return stub.result, stub.err
}

type enviarRepositoryStub struct {
	dados     application.OrcamentoParaEnvio
	enviadoEm time.Time
	errBusca  error
	errMarca  error
	usuarioID string
}

func (stub *enviarRepositoryStub) BuscarParaEnvio(context.Context, string) (application.OrcamentoParaEnvio, error) {
	return stub.dados, stub.errBusca
}

func (stub *enviarRepositoryStub) MarcarEnviado(_ context.Context, _, _, _, usuarioID string) (time.Time, error) {
	stub.usuarioID = usuarioID
	return stub.enviadoEm, stub.errMarca
}

type enviarNotificadorStub struct{ err error }

func (stub enviarNotificadorStub) Execute(context.Context, notificacaoApplication.Pedido) (notificacaoDominio.Notificacao, error) {
	return notificacaoDominio.Notificacao{}, stub.err
}

func TestConsultarHandler(t *testing.T) {
	const osID = "10000000-0000-0000-0000-000000000001"
	jwt, err := infrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	mecanico, _ := jwt.Gerar("mecanico", []string{"os:ler"})
	cliente, _ := jwt.GerarCliente("cliente", osID)
	clienteOutraOS, _ := jwt.GerarCliente("cliente", "20000000-0000-0000-0000-000000000001")
	estimativa := 3
	consulta := domain.Consulta{Cliente: domain.Cliente{ID: "cliente", Nome: "Ana", Documento: "123", TipoDocumento: "CPF"}, OrdemServicoID: osID, StatusOrdemServico: "EM_DIAGNOSTICO", Orcamentos: []domain.Orcamento{{ID: "orcamento", OriginalID: "original", Tipo: "PRINCIPAL", Status: "CRIADO", EstimativaDias: &estimativa, DataGeracao: time.Now(), Itens: []domain.Item{{Tipo: "SERVICO", Descricao: "Troca", Quantidade: 1, ValorUnitario: 100, ValorTotal: 100}}, Problemas: []domain.Problema{{ID: "problema", Descricao: "Ruido", RegistradoEm: time.Now()}}, ValorTotal: 100}}, ValorTotalGeral: 100}
	tests := []struct {
		name, id, token string
		err             error
		want            int
	}{
		{"sem token", osID, "", nil, http.StatusUnauthorized},
		{"id invalido", "invalido", mecanico, nil, http.StatusBadRequest},
		{"cliente outra os", osID, clienteOutraOS, nil, http.StatusForbidden},
		{"os ausente", osID, mecanico, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
		{"orcamento ausente", osID, mecanico, application.ErrOrcamentoNaoEncontrado, http.StatusNotFound},
		{"acesso negado", osID, mecanico, application.ErrAcessoNegado, http.StatusForbidden},
		{"erro interno", osID, mecanico, errors.New("falhou"), http.StatusInternalServerError},
		{"mecanico", osID, mecanico, nil, http.StatusOK},
		{"cliente", osID, cliente, nil, http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			useCase := application.NewConsultar(repositoryStub{result: consulta, err: test.err})
			mux.Handle("GET /ordens-servico/{osId}/orcamento", seguranca.RequireAnyScope(jwt, []string{"os:ler", "orcamentos:ler"}, NewConsultarHandler(useCase)))
			request := httptest.NewRequest(http.MethodGet, "/ordens-servico/"+test.id+"/orcamento", nil)
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
			if test.want == http.StatusOK && (!strings.Contains(writer.Body.String(), `"documento":"***.***.***-23"`) || strings.Contains(writer.Body.String(), `"documento":"123"`)) {
				t.Fatalf("documento exposto: %s", writer.Body.String())
			}
		})
	}
}

func TestAprovarHandler(t *testing.T) {
	const (
		orcamentoID = "10000000-0000-0000-0000-000000000001"
		osID        = "20000000-0000-0000-0000-000000000001"
		clienteID   = "30000000-0000-0000-0000-000000000001"
	)
	jwt, err := infrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	cliente, _ := jwt.GerarCliente(clienteID, osID)
	mecanicoID := "40000000-0000-0000-0000-000000000001"
	mecanico, _ := jwt.Gerar(mecanicoID, []string{"orcamentos:decidir"})
	result := domain.Aprovacao{
		OrcamentoID:        orcamentoID,
		OrdemServicoID:     osID,
		TipoOrcamento:      "PRINCIPAL",
		StatusOrcamento:    "APROVADO",
		StatusOrdemServico: "AGUARDANDO_EXECUCAO",
		ClienteID:          clienteID,
		DataAprovacao:      time.Now(),
	}
	tests := []struct {
		name, id, token, body string
		err                   error
		want                  int
	}{
		{"sem token", orcamentoID, "", "", nil, http.StatusUnauthorized},
		{"orcamento ausente", orcamentoID, cliente, "", application.ErrOrcamentoNaoEncontrado, http.StatusNotFound},
		{"acesso negado", orcamentoID, cliente, "", application.ErrAcessoNegado, http.StatusForbidden},
		{"conflito", orcamentoID, cliente, "", application.ErrOrcamentoJaDecidido, http.StatusConflict},
		{"ok", orcamentoID, cliente, "", nil, http.StatusOK},
		{"mecanico", orcamentoID, mecanico, "", nil, http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &aprovarStub{result: result, err: test.err}
			mux := http.NewServeMux()
			mux.Handle("POST /orcamentos/{orcamentoId}/aprovar", seguranca.RequireScope(jwt, "orcamentos:decidir", NewAprovarHandler(application.NewAprovar(stub, nil, nil))))
			request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+test.id+"/aprovar", strings.NewReader(test.body))
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
			if test.want == http.StatusOK && !strings.Contains(writer.Body.String(), `"statusOrcamento":"APROVADO"`) {
				t.Fatalf("body=%s", writer.Body.String())
			}
			if test.name == "mecanico" && (stub.input.ClienteID != "" || stub.input.UsuarioID != mecanicoID) {
				t.Fatalf("input=%+v", stub.input)
			}
		})
	}
}

func TestEnviarHandler(t *testing.T) {
	const (
		orcamentoID = "10000000-0000-0000-0000-000000000001"
		osID        = "20000000-0000-0000-0000-000000000001"
		usuarioID   = "30000000-0000-0000-0000-000000000001"
	)
	jwt, err := infrastructure.NewJWT("segredo")
	if err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.Gerar(usuarioID, []string{"orcamentos:escrever"})
	enviadoEm := time.Date(2026, time.August, 31, 9, 0, 0, 0, time.UTC)
	dados := application.OrcamentoParaEnvio{
		Orcamento: domain.Orcamento{
			ID:     orcamentoID,
			Tipo:   domain.TipoPrincipal,
			Status: domain.StatusCriado,
			Itens:  []domain.Item{{Tipo: "SERVICO", Descricao: "Troca", Quantidade: 1, ValorUnitario: 100}},
		},
		OrdemServicoID: osID,
		ClienteID:      "40000000-0000-0000-0000-000000000001",
		StatusOS:       "EM_DIAGNOSTICO",
		Calculado:      true,
		EstimativaDias: 2,
	}
	tests := []struct {
		name, id, token string
		err             error
		notificadorErr  error
		want            int
		body            string
	}{
		{"sem token", orcamentoID, "", nil, nil, http.StatusUnauthorized, ""},
		{"id invalido", "invalido", token, nil, nil, http.StatusBadRequest, ""},
		{"nao encontrado", orcamentoID, token, application.ErrOrcamentoNaoEncontrado, nil, http.StatusNotFound, ""},
		{"nao calculado", orcamentoID, token, domain.ErrOrcamentoNaoCalculado, nil, http.StatusConflict, ""},
		{"erro interno", orcamentoID, token, errors.New("falhou"), nil, http.StatusInternalServerError, ""},
		{"sucesso", orcamentoID, token, nil, nil, http.StatusOK, `"notificacaoEnviada":true`},
		{"notificacao falha", orcamentoID, token, nil, errors.New("smtp"), http.StatusOK, `"notificacaoEnviada":false`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &enviarRepositoryStub{dados: dados, enviadoEm: enviadoEm, errBusca: test.err}
			useCase := application.NewEnviar(repository, enviarNotificadorStub{err: test.notificadorErr}, log.New(io.Discard, "", 0))
			mux := http.NewServeMux()
			mux.Handle("POST /orcamentos/{orcamentoId}/enviar", seguranca.RequireScope(jwt, "orcamentos:escrever", NewEnviarHandler(useCase)))
			request := httptest.NewRequest(http.MethodPost, "/orcamentos/"+test.id+"/enviar", nil)
			if test.token != "" {
				request.Header.Set("Authorization", "Bearer "+test.token)
			}
			writer := httptest.NewRecorder()
			mux.ServeHTTP(writer, request)
			if writer.Code != test.want {
				t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
			}
			if test.body != "" && !strings.Contains(writer.Body.String(), test.body) {
				t.Fatalf("body=%s", writer.Body.String())
			}
			if test.name == "sucesso" && repository.usuarioID != usuarioID {
				t.Fatalf("usuarioID=%q", repository.usuarioID)
			}
		})
	}
}
