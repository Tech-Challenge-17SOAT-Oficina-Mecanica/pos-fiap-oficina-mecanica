package seguranca

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

type loginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}
type loginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"`
}

func NewLoginHandler(useCase seguranca.Autenticar) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var input loginRequest
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			problem(writer, 400, "Dados inválidos", "corpo da requisição inválido")
			return
		}
		token, err := useCase.Execute(request.Context(), input.Email, input.Senha)
		if err != nil {
			if errors.Is(err, seguranca.ErrDadosInvalidos) {
				problem(writer, 400, "Dados inválidos", err.Error())
				return
			}
			if errors.Is(err, seguranca.ErrCredenciaisInvalidas) {
				problem(writer, 401, "Não autorizado", err.Error())
				return
			}
			problem(writer, 500, "Erro interno", "falha ao autenticar mecânico")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(loginResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: 3600})
	}
}

func problem(writer http.ResponseWriter, status int, title, detail string) {
	sharedhttp.WriteProblem(writer, sharedhttp.Problem{Type: "https://api.oficina-mecanica.dev/errors/autenticacao", Title: title, Status: status, Detail: detail})
}
