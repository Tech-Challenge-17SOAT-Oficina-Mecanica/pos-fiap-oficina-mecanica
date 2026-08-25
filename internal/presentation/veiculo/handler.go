package veiculo

import (
	"encoding/json"
	"errors"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
	"io"
	"net/http"
	"strconv"
)

func NewInativarHandler(useCase application.Inativar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("veiculoId")
		if !validation.IsUUID(id) {
			writeProblem(w, 400, "Dados inválidos", "veiculoId inválido", "veiculoId")
			return
		}
		result, err := useCase.Execute(r.Context(), id, segurancaPresentation.UsuarioID(r.Context()), r.URL.Query().Get("motivo"))
		if errors.Is(err, application.ErrVeiculoJaInativo) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			if errors.As(err, &application.OSAbertaError{}) {
				writeProblem(w, 409, "Conflito de estado", err.Error(), "")
				return
			}
			status := 500
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				status = 404
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"veiculoId": result.Veiculo.ID, "placa": result.Veiculo.Placa, "ativo": false, "inativadoEm": result.Veiculo.InativadoEm, "inativadoPor": result.Veiculo.InativadoPor, "motivo": result.Veiculo.Motivo, "placaLiberadaParaNovoCadastro": true})
	}
}

func NewReativarHandler(useCase application.Reativar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("veiculoId")
		if !validation.IsUUID(id) {
			writeProblem(w, 400, "Dados inválidos", "veiculoId inválido", "veiculoId")
			return
		}
		result, err := useCase.Execute(r.Context(), id, segurancaPresentation.UsuarioID(r.Context()))
		if err != nil {
			status := 500
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				status = 404
			}
			if errors.Is(err, application.ErrVeiculoJaAtivo) || errors.Is(err, application.ErrPlacaDuplicada) || errors.Is(err, application.ErrClienteProprietarioInativo) {
				status = 409
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"veiculoId": result.Veiculo.ID, "placa": result.Veiculo.Placa, "ativo": true, "reativadoEm": result.ReativadoEm, "reativadoPor": result.ReativadoPor})
	}
}

type request struct {
	Placa  string `json:"placa"`
	Marca  string `json:"marca"`
	Modelo string `json:"modelo"`
	Ano    int    `json:"ano"`
}

func NewAtualizarHandler(useCase application.Atualizar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("veiculoId")
		if !validation.IsUUID(id) {
			writeProblem(w, 400, "Dados inválidos", "veiculoId inválido", "veiculoId")
			return
		}
		ifMatch := r.Header.Get("If-Match")
		if ifMatch == "" {
			writeProblem(w, 428, "Pré-condição obrigatória", "If-Match é obrigatório", "")
			return
		}
		version, err := strconv.Atoi(ifMatch)
		if err != nil || version < 1 {
			writeProblem(w, 400, "Dados inválidos", "If-Match inválido", "If-Match")
			return
		}
		var input request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(w, 400, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		cadastro, err := domain.NovoCadastro(input.Placa, input.Marca, input.Modelo, input.Ano)
		if err != nil {
			writeProblem(w, 400, "Dados inválidos", err.Error(), "veiculo")
			return
		}
		v, err := useCase.Execute(r.Context(), id, version, cadastro)
		if err != nil {
			status := 500
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				status = 404
			}
			if errors.Is(err, application.ErrPlacaDuplicada) {
				status = 409
			}
			if errors.Is(err, application.ErrVersaoDivergente) {
				status = 412
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

func NewConsultaHandler(useCase application.Consultar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pl, err := domain.NormalizarPlaca(r.URL.Query().Get("placa"))
		if err != nil {
			writeProblem(w, 400, "Dados inválidos", err.Error(), "placa")
			return
		}
		incluir := r.URL.Query().Get("incluirInativos")
		if incluir != "true" && incluir != "false" {
			writeProblem(w, 400, "Dados inválidos", "incluirInativos inválido", "incluirInativos")
			return
		}
		v, err := useCase.Execute(r.Context(), pl, incluir == "true")
		if err != nil {
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				writeProblem(w, 404, titleFor(404), err.Error(), "")
				return
			}
			writeProblem(w, 500, titleFor(500), "falha ao consultar veículo", "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

func NewHandler(useCase application.Cadastrar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clienteID := r.PathValue("clienteId")
		if !validation.IsUUID(clienteID) {
			writeProblem(w, 400, "Dados inválidos", "clienteId inválido", "clienteId")
			return
		}
		var input request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(w, 400, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		cadastro, err := domain.NovoCadastro(input.Placa, input.Marca, input.Modelo, input.Ano)
		if err != nil {
			writeProblem(w, 400, "Dados inválidos", err.Error(), "veiculo")
			return
		}
		v, err := useCase.Execute(r.Context(), clienteID, cadastro)
		if err != nil {
			status := 500
			if errors.Is(err, application.ErrClienteNaoEncontrado) {
				status = 404
			}
			if errors.Is(err, application.ErrClienteInativo) || errors.Is(err, application.ErrPlacaDuplicada) {
				status = 409
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(v)
	}
}
func writeProblem(w http.ResponseWriter, status int, title, detail, campo string) {
	problem := sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/veiculo", Title: title, Status: status, Detail: detail}
	if campo != "" {
		problem.Erros = []sharedhttp.FieldError{{Campo: campo, Mensagem: detail}}
	}
	sharedhttp.WriteProblem(w, problem)
}

func titleFor(status int) string {
	if status == http.StatusNotFound {
		return "Recurso não encontrado"
	}
	if status == http.StatusConflict {
		return "Conflito de estado"
	}
	return "Erro interno"
}
