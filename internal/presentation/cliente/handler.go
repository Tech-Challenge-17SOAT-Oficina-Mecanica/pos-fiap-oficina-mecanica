package cliente

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
	seguranca "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const escopoCadastrarCliente = "clientes:escrever"
const escopoConsultarCliente = "clientes:ler"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type CadastrarUseCase interface {
	Execute(context.Context, domain.NovoClienteInput) (domain.Cliente, error)
}

type ConsultarUseCase interface {
	Execute(context.Context, string) (domain.Cliente, error)
}

type AtualizarUseCase interface {
	Execute(context.Context, application.AtualizarInput) (domain.Cliente, error)
}

type InativarUseCase interface {
	Execute(context.Context, application.InativarInput) (application.Inativacao, error)
}

type ReativarUseCase interface {
	Execute(context.Context, application.ReativarInput) (application.Reativacao, error)
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

type atualizarClienteResponse struct {
	ID            string `json:"id"`
	Nome          string `json:"nome"`
	Documento     string `json:"documento"`
	TipoDocumento string `json:"tipoDocumento"`
	Telefone      string `json:"telefone,omitempty"`
	Email         string `json:"email,omitempty"`
	Ativo         bool   `json:"ativo"`
	Version       int    `json:"version"`
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

type inativarResponse struct {
	ClienteID                         string                    `json:"clienteId"`
	Nome                              string                    `json:"nome"`
	Ativo                             bool                      `json:"ativo"`
	InativadoEm                       string                    `json:"inativadoEm"`
	InativadoPor                      string                    `json:"inativadoPor"`
	Motivo                            string                    `json:"motivo,omitempty"`
	VeiculosInativados                []domain.VeiculoInativado `json:"veiculosInativados"`
	DocumentoLiberadoParaNovoCadastro bool                      `json:"documentoLiberadoParaNovoCadastro"`
}

type reativarResponse struct {
	ClienteID          string `json:"clienteId"`
	Nome               string `json:"nome"`
	Ativo              bool   `json:"ativo"`
	ReativadoEm        string `json:"reativadoEm"`
	ReativadoPor       string `json:"reativadoPor"`
	VeiculosReativados int    `json:"veiculosReativados"`
}

func NewCadastrarHandler(useCase CadastrarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if _, ok := autorizado(writer, request, token, escopoCadastrarCliente); !ok {
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
		if _, ok := autorizado(writer, request, token, escopoConsultarCliente); !ok {
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

func NewAtualizarHandler(useCase AtualizarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if _, ok := autorizado(writer, request, token, escopoCadastrarCliente); !ok {
			return
		}
		ifMatch := strings.TrimSpace(request.Header.Get("If-Match"))
		if ifMatch == "" {
			problem(writer, http.StatusPreconditionRequired, "Pré-condição obrigatória", "If-Match não informado")
			return
		}
		version, err := strconv.Atoi(ifMatch)
		if err != nil || version <= 0 {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "If-Match inválido")
			return
		}
		var input cadastrarRequest
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido")
			return
		}
		cliente, err := useCase.Execute(request.Context(), application.AtualizarInput{
			ClienteID: request.PathValue("clienteId"),
			Version:   version,
			Dados: domain.AtualizarClienteInput{
				Nome:          input.Nome,
				Documento:     input.Documento,
				TipoDocumento: input.TipoDocumento,
				Telefone:      input.Telefone,
				Email:         input.Email,
			},
		})
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(atualizarClienteResponse{
			ID:            cliente.ID,
			Nome:          cliente.Nome,
			Documento:     cliente.Documento,
			TipoDocumento: cliente.TipoDocumento,
			Telefone:      cliente.Telefone,
			Email:         cliente.Email,
			Ativo:         cliente.Ativo,
			Version:       cliente.Version,
		})
	}
}

func NewInativarHandler(useCase InativarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		claims, ok := autorizado(writer, request, token, escopoCadastrarCliente)
		if !ok {
			return
		}
		clienteID := request.PathValue("clienteId")
		if !uuidValido(clienteID) {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "clienteId inválido")
			return
		}
		result, err := useCase.Execute(request.Context(), application.InativarInput{
			ClienteID: clienteID,
			UsuarioID: claims.UsuarioID,
			Motivo:    request.URL.Query().Get("motivo"),
		})
		if errors.Is(err, application.ErrClienteJaInativo) {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(inativarResponse{
			ClienteID:                         result.Cliente.ID,
			Nome:                              result.Cliente.Nome,
			Ativo:                             result.Cliente.Ativo,
			InativadoEm:                       result.Cliente.InativadoEm.Format("2006-01-02T15:04:05Z07:00"),
			InativadoPor:                      result.Cliente.InativadoPor,
			Motivo:                            result.Cliente.Motivo,
			VeiculosInativados:                result.VeiculosInativados,
			DocumentoLiberadoParaNovoCadastro: result.DocumentoLiberado,
		})
	}
}

func NewReativarHandler(useCase ReativarUseCase, token TokenValidator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		claims, ok := autorizado(writer, request, token, escopoCadastrarCliente)
		if !ok {
			return
		}
		clienteID := request.PathValue("clienteId")
		if !uuidValido(clienteID) {
			problem(writer, http.StatusBadRequest, "Dados inválidos", "clienteId inválido")
			return
		}
		result, err := useCase.Execute(request.Context(), application.ReativarInput{ClienteID: clienteID, UsuarioID: claims.UsuarioID})
		if err != nil {
			writeError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(reativarResponse{
			ClienteID:          result.Cliente.ID,
			Nome:               result.Cliente.Nome,
			Ativo:              result.Cliente.Ativo,
			ReativadoEm:        result.ReativadoEm.Format("2006-01-02T15:04:05Z07:00"),
			ReativadoPor:       claims.UsuarioID,
			VeiculosReativados: result.VeiculosReativados,
		})
	}
}

func autorizado(writer http.ResponseWriter, request *http.Request, token TokenValidator, escopoExigido string) (seguranca.Claims, bool) {
	raw := strings.TrimSpace(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "))
	if raw == "" || raw == request.Header.Get("Authorization") {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return seguranca.Claims{}, false
	}
	claims, err := token.Validar(raw)
	if err != nil {
		problem(writer, http.StatusUnauthorized, "Não autorizado", "token ausente ou expirado")
		return seguranca.Claims{}, false
	}
	for _, escopo := range claims.Escopos {
		if escopo == escopoExigido {
			return claims, true
		}
	}
	problem(writer, http.StatusForbidden, "Acesso negado", "usuário sem o escopo "+escopoExigido)
	return seguranca.Claims{}, false
}

func writeError(writer http.ResponseWriter, err error) {
	var osAbertas application.OSAbertaError
	if errors.As(err, &osAbertas) {
		sharedhttp.WriteProblem(writer, sharedhttp.Problem{
			Type:   "https://api.oficina-mecanica.dev/errors/cliente-com-os-aberta",
			Title:  "Cliente possui Ordem de Serviço em aberto",
			Status: http.StatusConflict,
			Detail: "Não é possível excluir o cliente enquanto houver OS em andamento",
			Erros:  osAbertas.Ordens,
		})
		return
	}
	if errors.Is(err, application.ErrClienteDuplicado) {
		problem(writer, http.StatusConflict, "Conflito", err.Error())
		return
	}
	if errors.Is(err, application.ErrClienteNaoEncontrado) {
		problem(writer, http.StatusNotFound, "Não encontrado", err.Error())
		return
	}
	if errors.Is(err, application.ErrVersaoDivergente) {
		problem(writer, http.StatusPreconditionFailed, "Pré-condição falhou", err.Error())
		return
	}
	if errors.Is(err, application.ErrClienteJaAtivo) {
		problem(writer, http.StatusConflict, "Conflito", err.Error())
		return
	}
	if errClienteInvalido(err) {
		problem(writer, http.StatusBadRequest, "Dados inválidos", err.Error())
		return
	}
	problem(writer, http.StatusInternalServerError, "Erro interno", "falha ao processar cliente")
}

func errClienteInvalido(err error) bool {
	return errors.Is(err, domain.ErrNomeObrigatorio) ||
		errors.Is(err, domain.ErrDocumentoObrigatorio) ||
		errors.Is(err, domain.ErrTipoDocumentoObrigatorio) ||
		errors.Is(err, domain.ErrTipoDocumentoInvalido) ||
		errors.Is(err, domain.ErrDocumentoInvalido) ||
		errors.Is(err, domain.ErrContatoObrigatorio) ||
		errors.Is(err, domain.ErrTelefoneInvalido) ||
		errors.Is(err, domain.ErrEmailInvalido) ||
		errors.Is(err, domain.ErrClienteIDObrigatorio) ||
		errors.Is(err, domain.ErrMotivoInvalido)
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/clientes", Title: title, Status: status, Detail: detail})
}

func uuidValido(value string) bool {
	return uuidPattern.MatchString(value)
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
