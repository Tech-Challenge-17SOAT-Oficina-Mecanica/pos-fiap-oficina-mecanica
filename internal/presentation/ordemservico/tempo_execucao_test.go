package ordemservico

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type tempoExecucaoRepositoryStub struct {
	item  domain.TempoExecucao
	itens []domain.TempoExecucao
	total int
	soma  int64
	err   error
}

func (stub tempoExecucaoRepositoryStub) ConsultarTempoExecucao(context.Context, string) (domain.TempoExecucao, error) {
	return stub.item, stub.err
}

func (stub tempoExecucaoRepositoryStub) ListarTemposExecucao(context.Context, *time.Time, *time.Time, int, int) ([]domain.TempoExecucao, int, int64, error) {
	return stub.itens, stub.total, stub.soma, stub.err
}

func TestTempoExecucaoHandlers(t *testing.T) {
	const osID = "70000000-0000-0000-0000-000000000001"
	inicio := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	item, err := domain.NovoTempoExecucao(osID, inicio, inicio.Add(90*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("consulta por os", func(t *testing.T) {
		for _, test := range []struct {
			name string
			id   string
			err  error
			want int
		}{
			{"id invalido", "invalido", nil, http.StatusBadRequest},
			{"nao encontrada", osID, application.ErrOrdemServicoNaoEncontrada, http.StatusNotFound},
			{"tempo indisponivel", osID, domain.ErrTempoExecucaoIndisponivel, http.StatusBadRequest},
			{"erro interno", osID, errors.New("db"), http.StatusInternalServerError},
			{"sucesso", osID, nil, http.StatusOK},
		} {
			t.Run(test.name, func(t *testing.T) {
				mux := http.NewServeMux()
				mux.Handle("GET /ordens-servico/{osId}/tempo-execucao", NewConsultarTempoExecucaoHandler(application.NewConsultarTempoExecucaoDaOS(tempoExecucaoRepositoryStub{item: item, err: test.err})))
				writer := httptest.NewRecorder()
				mux.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/ordens-servico/"+test.id+"/tempo-execucao", nil))
				if writer.Code != test.want {
					t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
				}
				if test.want == http.StatusOK && !strings.Contains(writer.Body.String(), `"tempoExecucaoMinutos":90`) {
					t.Fatalf("resposta invalida: %s", writer.Body.String())
				}
			})
		}
	})

	t.Run("lista tempos", func(t *testing.T) {
		for _, test := range []struct {
			name string
			url  string
			err  error
			want int
		}{
			{"paginacao invalida", "/ordens-servico/tempos-execucao?tamanho=999", nil, http.StatusBadRequest},
			{"data inicio invalida", "/ordens-servico/tempos-execucao?dataInicio=x", application.ErrDataInicioInvalida, http.StatusBadRequest},
			{"data fim invalida", "/ordens-servico/tempos-execucao?dataFim=x", application.ErrDataFimInvalida, http.StatusBadRequest},
			{"periodo invalido", "/ordens-servico/tempos-execucao?dataInicio=2026-09-01&dataFim=2026-08-01", nil, http.StatusBadRequest},
			{"erro interno", "/ordens-servico/tempos-execucao", errors.New("db"), http.StatusInternalServerError},
			{"sucesso", "/ordens-servico/tempos-execucao?tamanho=10&pagina=0", nil, http.StatusOK},
		} {
			t.Run(test.name, func(t *testing.T) {
				mux := http.NewServeMux()
				mux.Handle("GET /ordens-servico/tempos-execucao", NewListarTemposExecucaoHandler(application.NewConsultarTempoMedioExecucao(tempoExecucaoRepositoryStub{
					itens: []domain.TempoExecucao{item}, total: 1, soma: 90, err: test.err,
				})))
				writer := httptest.NewRecorder()
				mux.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, test.url, nil))
				if writer.Code != test.want {
					t.Fatalf("status=%d body=%s", writer.Code, writer.Body.String())
				}
				if test.want == http.StatusOK && !strings.Contains(writer.Body.String(), `"tempoMedioExecucaoMinutos":90`) {
					t.Fatalf("resposta invalida: %s", writer.Body.String())
				}
			})
		}
	})
}
