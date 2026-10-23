package veiculo

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

func NewInativarHandler(useCase application.Inativar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("veiculoId")
		if !validation.IsUUID(id) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "veiculoId inválido", "veiculoId")
			return
		}
		resultado, err := useCase.Execute(r.Context(), id, segurancaPresentation.UsuarioID(r.Context()), r.URL.Query().Get("motivo"))
		if errors.Is(err, application.ErrVeiculoJaInativo) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			if errors.As(err, &application.OSAbertaError{}) {
				writeProblem(w, http.StatusConflict, "Conflito de estado", err.Error(), "")
				return
			}
			status := http.StatusInternalServerError
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				status = http.StatusNotFound
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"veiculoId": resultado.Veiculo.ID, "placa": resultado.Veiculo.Placa, "ativo": false, "inativadoEm": resultado.Veiculo.InativadoEm, "inativadoPor": resultado.Veiculo.InativadoPor, "motivo": resultado.Veiculo.Motivo, "placaLiberadaParaNovoCadastro": true})
	}
}

func NewReativarHandler(useCase application.Reativar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("veiculoId")
		if !validation.IsUUID(id) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "veiculoId inválido", "veiculoId")
			return
		}
		resultado, err := useCase.Execute(r.Context(), id, segurancaPresentation.UsuarioID(r.Context()))
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				status = http.StatusNotFound
			}
			if errors.Is(err, application.ErrVeiculoJaAtivo) || errors.Is(err, application.ErrPlacaDuplicada) || errors.Is(err, application.ErrClienteProprietarioInativo) {
				status = http.StatusConflict
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"veiculoId": resultado.Veiculo.ID, "placa": resultado.Veiculo.Placa, "ativo": true, "reativadoEm": resultado.ReativadoEm, "reativadoPor": resultado.ReativadoPor})
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
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "veiculoId inválido", "veiculoId")
			return
		}
		ifMatch := r.Header.Get("If-Match")
		if ifMatch == "" {
			writeProblem(w, http.StatusPreconditionRequired, "Pré-condição obrigatória", "If-Match é obrigatório", "")
			return
		}
		version, err := strconv.Atoi(ifMatch)
		if err != nil || version < 1 {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "If-Match inválido", "If-Match")
			return
		}
		var input request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		cadastro, err := domain.NovoCadastro(input.Placa, input.Marca, input.Modelo, input.Ano)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "veiculo")
			return
		}
		veiculo, err := useCase.Execute(r.Context(), id, version, cadastro)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				status = http.StatusNotFound
			}
			if errors.Is(err, application.ErrPlacaDuplicada) {
				status = http.StatusConflict
			}
			if errors.Is(err, application.ErrVersaoDivergente) {
				status = http.StatusPreconditionFailed
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(veiculo)
	}
}

func NewConsultaHandler(useCase application.Consultar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		placa, err := domain.NormalizarPlaca(r.URL.Query().Get("placa"))
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "placa")
			return
		}
		incluirInativos := r.URL.Query().Get("incluirInativos")
		if incluirInativos != "true" && incluirInativos != "false" {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "incluirInativos inválido", "incluirInativos")
			return
		}
		veiculo, err := useCase.Execute(r.Context(), placa, incluirInativos == "true")
		if err != nil {
			if errors.Is(err, application.ErrVeiculoNaoEncontrado) {
				writeProblem(w, http.StatusNotFound, titleFor(http.StatusNotFound), err.Error(), "")
				return
			}
			writeProblem(w, http.StatusInternalServerError, titleFor(http.StatusInternalServerError), "falha ao consultar veículo", "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(veiculo)
	}
}

func NewCadastrarHandler(useCase application.Cadastrar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clienteID := r.PathValue("clienteId")
		if !validation.IsUUID(clienteID) {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "clienteId inválido", "clienteId")
			return
		}
		var input request
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", "corpo da requisição inválido", "")
			return
		}
		cadastro, err := domain.NovoCadastro(input.Placa, input.Marca, input.Modelo, input.Ano)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Dados inválidos", err.Error(), "veiculo")
			return
		}
		veiculo, err := useCase.Execute(r.Context(), clienteID, cadastro)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, application.ErrClienteNaoEncontrado) {
				status = http.StatusNotFound
			}
			if errors.Is(err, application.ErrClienteInativo) || errors.Is(err, application.ErrPlacaDuplicada) {
				status = http.StatusConflict
			}
			writeProblem(w, status, titleFor(status), err.Error(), "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(veiculo)
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
