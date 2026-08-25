package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	clienteApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	pecaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	clientePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	pecaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const (
	escopoEstoqueLer      = "estoque:ler"
	escopoEstoqueEscrever = "estoque:escrever"
)

func main() {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	jwt, err := segurancaInfrastructure.NewJWT(os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}
	login := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt)

	pecaRepository := pecaInfrastructure.NewPostgresRepository(db)
	consultarPecas := pecaApplication.NewConsultarPecas(pecaRepository)
	desativarPeca := pecaApplication.NewDesativarPeca(pecaRepository)

	clienteRepository := clienteInfrastructure.NewPostgresRepository(db)
	cadastrarCliente := clienteApplication.NewCadastrar(clienteRepository)
	consultarCliente := clienteApplication.NewConsultar(clienteRepository)
	atualizarCliente := clienteApplication.NewAtualizar(clienteRepository)
	inativarCliente := clienteApplication.NewInativar(clienteRepository)
	reativarCliente := clienteApplication.NewReativar(clienteRepository)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("GET /clientes", clientePresentation.NewConsultarHandler(consultarCliente, jwt))
	mux.Handle("POST /clientes", clientePresentation.NewCadastrarHandler(cadastrarCliente, jwt))
	mux.Handle("PUT /clientes/{clienteId}", clientePresentation.NewAtualizarHandler(atualizarCliente, jwt))
	mux.Handle("DELETE /clientes/{clienteId}", clientePresentation.NewInativarHandler(inativarCliente, jwt))
	mux.Handle("POST /clientes/{clienteId}/reativacao", clientePresentation.NewReativarHandler(reativarCliente, jwt))
	mux.Handle("GET /estoque/pecas", segurancaPresentation.ComEscopo(jwt, escopoEstoqueLer,
		pecaPresentation.NewConsultarPecasHandler(consultarPecas)))
	mux.Handle("GET /estoque/pecas/{pecaId}", segurancaPresentation.ComEscopo(jwt, escopoEstoqueLer,
		pecaPresentation.NewConsultarPecaPorIDHandler(consultarPecas)))
	mux.Handle("DELETE /estoque/pecas/{pecaId}", segurancaPresentation.ComEscopo(jwt, escopoEstoqueEscrever,
		pecaPresentation.NewDesativarPecaHandler(desativarPeca)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: sharedhttp.CORS(mux),
	}

	log.Println("API iniciada na porta 8080")
	log.Fatal(server.ListenAndServe())
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(map[string]string{"status": "ok"})
}
