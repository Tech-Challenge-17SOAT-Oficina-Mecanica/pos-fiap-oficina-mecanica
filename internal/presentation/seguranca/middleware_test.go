package seguranca

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type autenticadorFake struct {
	usuarioID string
	escopos   []string
	erro      error
}

func (fake autenticadorFake) Autenticar(string) (string, []string, error) {
	return fake.usuarioID, fake.escopos, fake.erro
}

func TestComEscopo(t *testing.T) {
	casos := []struct {
		nome          string
		authorization string
		autenticador  autenticadorFake
		status        int
		chamaProximo  bool
	}{
		{"sem header", "", autenticadorFake{}, http.StatusUnauthorized, false},
		{"esquema errado", "Basic abc", autenticadorFake{}, http.StatusUnauthorized, false},
		{"bearer vazio", "Bearer   ", autenticadorFake{}, http.StatusUnauthorized, false},
		{"token invalido", "Bearer x", autenticadorFake{erro: errors.New("expirado")}, http.StatusUnauthorized, false},
		{"sem o escopo", "Bearer x", autenticadorFake{usuarioID: "u1", escopos: []string{"clientes:ler"}}, http.StatusForbidden, false},
		{"com o escopo", "Bearer x", autenticadorFake{usuarioID: "u1", escopos: []string{"estoque:ler"}}, http.StatusOK, true},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			chamado := false
			protegido := ComEscopo(caso.autenticador, "estoque:ler", func(writer http.ResponseWriter, request *http.Request) {
				chamado = true
				if UsuarioID(request.Context()) != caso.autenticador.usuarioID {
					t.Fatalf("usuarioID no contexto = %q", UsuarioID(request.Context()))
				}
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/estoque/pecas", nil)
			if caso.authorization != "" {
				request.Header.Set("Authorization", caso.authorization)
			}
			protegido(response, request)

			if response.Code != caso.status {
				t.Fatalf("status = %d, esperado %d", response.Code, caso.status)
			}
			if chamado != caso.chamaProximo {
				t.Fatalf("handler seguinte chamado = %v, esperado %v", chamado, caso.chamaProximo)
			}
		})
	}
}
