package veiculo

import (
	"encoding/json"
	"errors"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
	"io"
	"net/http"
)

type request struct {
	Placa  string `json:"placa"`
	Marca  string `json:"marca"`
	Modelo string `json:"modelo"`
	Ano    int    `json:"ano"`
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
