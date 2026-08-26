package peca

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	pecaDomain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

type desativarRepositorioFake struct {
	encontrada  pecaDomain.Peca
	erroBusca   error
	ordens      []string
	emOrcamento bool
}

func (fake desativarRepositorioFake) BuscarPorID(context.Context, string) (pecaDomain.Peca, error) {
	return fake.encontrada, fake.erroBusca
}

func (fake desativarRepositorioFake) OrdensComReservaAtiva(context.Context, string) ([]string, error) {
	return fake.ordens, nil
}

func (fake desativarRepositorioFake) EmOrcamentoCriado(context.Context, string) (bool, error) {
	return fake.emOrcamento, nil
}

func (fake desativarRepositorioFake) Desativar(context.Context, pecaDomain.Peca) error {
	return nil
}

func desativar(t *testing.T, pecaID string, fake desativarRepositorioFake) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest("DELETE", "/estoque/pecas/"+pecaID, nil)
	request.SetPathValue("pecaId", pecaID)
	NewDesativarPecaHandler(peca.NewDesativarPeca(fake))(response, request)
	return response
}

func pecaAtivaExemplo() pecaDomain.Peca {
	return pecaDomain.Peca{
		ID:     "50000000-0000-0000-0000-000000000001",
		Codigo: "PEC-000001",
		Nome:   "Filtro de oleo",
		Ativo:  true,
	}
}

func TestDesativarRetorna200ComRecursoAtualizado(t *testing.T) {
	response := desativar(t, "50000000-0000-0000-0000-000000000001",
		desativarRepositorioFake{encontrada: pecaAtivaExemplo()})

	if response.Code != 200 {
		t.Fatalf("status = %d, corpo %s", response.Code, response.Body)
	}

	var corpo struct {
		ID                 string  `json:"id"`
		Codigo             string  `json:"codigo"`
		Nome               string  `json:"nome"`
		Ativo              bool    `json:"ativo"`
		DataDesativacao    *string `json:"dataDesativacao"`
		UsuarioDesativacao *string `json:"usuarioDesativacao"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &corpo); err != nil {
		t.Fatal(err)
	}
	if corpo.Ativo {
		t.Fatalf("ativo deveria ser false: %+v", corpo)
	}
	if corpo.Codigo != "PEC-000001" || corpo.Nome != "Filtro de oleo" {
		t.Fatalf("recurso atualizado incompleto: %+v", corpo)
	}
	if corpo.DataDesativacao == nil {
		t.Fatalf("dataDesativacao ausente: %s", response.Body)
	}
	if strings.Contains(response.Body.String(), `"data"`) {
		t.Fatalf("recurso unico nao deve usar envelope: %s", response.Body)
	}
}

func TestDesativarMapeiaErrosParaStatus(t *testing.T) {
	casos := []struct {
		nome   string
		pecaID string
		fake   desativarRepositorioFake
		status int
	}{
		{
			"identificador invalido", "abc",
			desativarRepositorioFake{encontrada: pecaAtivaExemplo()}, 400,
		},
		{
			"peca inexistente", "50000000-0000-0000-0000-0000000000ff",
			desativarRepositorioFake{erroBusca: peca.ErrNaoEncontrada}, 404,
		},
		{
			"peca ja inativa", "50000000-0000-0000-0000-000000000001",
			desativarRepositorioFake{encontrada: pecaDomain.Peca{ID: "x", Ativo: false}}, 409,
		},
		{
			"saldo reservado", "50000000-0000-0000-0000-000000000001",
			desativarRepositorioFake{encontrada: pecaAtivaExemplo(), ordens: []string{"os-1"}}, 409,
		},
		{
			"em orcamento criado", "50000000-0000-0000-0000-000000000001",
			desativarRepositorioFake{encontrada: pecaAtivaExemplo(), emOrcamento: true}, 409,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			response := desativar(t, caso.pecaID, caso.fake)
			if response.Code != caso.status {
				t.Fatalf("status = %d, esperado %d, corpo %s", response.Code, caso.status, response.Body)
			}
			if response.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("Content-Type = %q", response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestDesativarInformaAsOrdensQueSeguramAReserva(t *testing.T) {
	response := desativar(t, "50000000-0000-0000-0000-000000000001",
		desativarRepositorioFake{encontrada: pecaAtivaExemplo(), ordens: []string{"os-1", "os-2"}})

	if !strings.Contains(response.Body.String(), "os-1") || !strings.Contains(response.Body.String(), "os-2") {
		t.Fatalf("detail deve listar as OS: %s", response.Body)
	}
}
