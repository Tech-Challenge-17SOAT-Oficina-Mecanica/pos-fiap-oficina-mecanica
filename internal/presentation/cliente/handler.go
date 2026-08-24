package cliente

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const escopoCadastrarCliente = "clientes:escrever"
const escopoConsultarCliente = "clientes:ler"

type CadastrarUseCase interface {
	Execute(context.Context, domain.NovoClienteInput) (domain.Cliente, error)
}

type ConsultarUseCase interface {
	Execute(context.Context, string) (domain.Cliente, error)
}

type TokenValidator interface {
	Validar(string) (seguranca.Claims, error)
}

type cadastrarRequest struct {
	Nome          string `json:"nome"`
	Documento     string `json:"documento"`
	TipoDocumento string `json:"tipoDocumento"`
	Telefone      string `json:"telefone"`
	Email         string `json:"email"`
}

type clienteResponse struct {
	ID            string `json:"id"`
	Nome          string `json:"nome"`
	Documento     string `json:"documento"`
	TipoDocumento string `json:"tipoDocumento"`
	Telefone      string `json:"telefone,omitempty"`
	Email         string `json:"email,omitempty"`
}

type consultaClienteResponse struct {
	ID            string            `json:"id"`
	Nome          string            `json:"nome"`
	Documento     string            `json:"documento"`
	TipoDocumento string            `json:"tipoDocumento"`
	Telefone      string            `json:"telefone,omitempty"`
	Email         string            `json:"email,omitempty"`
	Ativo         bool              `json:"ativo"`
	Version       int               `json:"version"`
	Veiculos      []veiculoResponse `json:"veiculos"`
}

type veiculoResponse struct {
	ID     string `json:"id"`
	Placa  string `json:"placa"`
	Marca  string `json:"marca"`
	Modelo string `json:"modelo"`
	Ano    int    `json:"ano"`
}

func NewCadastrarHandler(useCase CadastrarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !autorizado(writer, request, token, escopoCadastrarCliente) {
			return
		}
		var input cadastrarRequest
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido")
			return
		}
		cliente, err := useCase.Execute(request.Context(), domain.NovoClienteInput{
			Nome:          input.Nome,
			Documento:     input.Documento,
			TipoDocumento: input.TipoDocumento,
			Telefone:      input.Telefone,
			Email:         input.Email,
		})
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(writer).Encode(clienteResponse{
			ID:            cliente.ID,
			Nome:          cliente.Nome,
			Documento:     cliente.Documento,
			TipoDocumento: cliente.TipoDocumento,
			Telefone:      cliente.Telefone,
			Email:         cliente.Email,
		})
	}
}

func NewConsultarHandler(useCase ConsultarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if !autorizado(writer, request, token, escopoConsultarCliente) {
			return
		}
		cliente, err := useCase.Execute(request.Context(), request.URL.Query().Get("documento"))
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(toConsultaResponse(cliente))
	}
}

func autorizado(writer http.ResponseWriter, request *http.Request, token TokenValidator, escopoExigido string) bool {
	raw := strings.TrimSpace(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "))
	if raw == "" || raw == request.Header.Get("Authorization") {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return false
	}
	claims, err := token.Validar(raw)
	if err != nil {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return false
	}
	for _, escopo := range claims.Escopos {
		if escopo == escopoExigido {
			return true
		}
	}
	problem(writer, http.StatusForbidden, "Acesso negado", "usuário sem o escopo "+escopoExigido)
	return false
}

func writeError(writer http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrClienteDuplicado) {
		problem(writer, http.StatusConflict, "Conflito", err.Error())
		return
	}
	if errors.Is(err, application.ErrClienteNaoEncontrado) {
		problem(writer, http.StatusNotFound, "Não encontrado", err.Error())
		return
	}
	if errClienteInvalido(err) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
		return
	}
	problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao cadastrar cliente")
}

func errClienteInvalido(err error) bool {
	return errors.Is(err, domain.ErrNomeObrigatorio) ||
		errors.Is(err, domain.ErrDocumentoObrigatorio) ||
		errors.Is(err, domain.ErrTipoDocumentoObrigatorio) ||
		errors.Is(err, domain.ErrTipoDocumentoInvalido) ||
		errors.Is(err, domain.ErrDocumentoInvalido) ||
		errors.Is(err, domain.ErrContatoObrigatorio) ||
		errors.Is(err, domain.ErrTelefoneInvalido) ||
		errors.Is(err, domain.ErrEmailInvalido)
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/clientes", Title: title, Status: status, Detail: detail})
}

func toConsultaResponse(cliente domain.Cliente) consultaClienteResponse {
	response := consultaClienteResponse{
		ID:            cliente.ID,
		Nome:          cliente.Nome,
		Documento:     cliente.Documento,
		TipoDocumento: cliente.TipoDocumento,
		Telefone:      cliente.Telefone,
		Email:         cliente.Email,
		Ativo:         cliente.Ativo,
		Version:       cliente.Version,
		Veiculos:      make([]veiculoResponse, 0, len(cliente.Veiculos)),
	}
	for _, veiculo := range cliente.Veiculos {
		response.Veiculos = append(response.Veiculos, veiculoResponse{
			ID:     veiculo.ID,
			Placa:  veiculo.Placa,
			Marca:  veiculo.Marca,
			Modelo: veiculo.Modelo,
			Ano:    veiculo.Ano,
		})
	}
	return response
}
