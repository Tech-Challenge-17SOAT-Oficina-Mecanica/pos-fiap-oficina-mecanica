package main

import (
	"log"
	"net/http"
	"os"

	clienteApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	veiculoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	clientePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	veiculoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	clienteDB, err := database.OpenPool()
	if err != nil {
		log.Fatal(err)
	}
	defer clienteDB.Close()
	jwt, err := segurancaInfrastructure.NewJWT(os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	login := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt)
	cadastrar := veiculoApplication.NewCadastrar(veiculo.NewPostgresRepository(db))
	consultar := veiculoApplication.NewConsultar(veiculo.NewPostgresRepository(db))
	atualizar := veiculoApplication.NewAtualizar(veiculo.NewPostgresRepository(db))
	clienteRepository := clienteInfrastructure.NewPostgresRepository(clienteDB)
	cadastrarCliente := clienteApplication.NewCadastrar(clienteRepository)
	consultarCliente := clienteApplication.NewConsultar(clienteRepository)
	atualizarCliente := clienteApplication.NewAtualizar(clienteRepository)
	inativarCliente := clienteApplication.NewInativar(clienteRepository)
	reativarCliente := clienteApplication.NewReativar(clienteRepository)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("POST /clientes/{clienteId}/veiculos", segurancaPresentation.RequireScope(jwt, "veiculos:escrever", veiculoPresentation.NewHandler(cadastrar)))
	mux.Handle("GET /veiculos", segurancaPresentation.RequireScope(jwt, "veiculos:ler", veiculoPresentation.NewConsultaHandler(consultar)))
	mux.Handle("PUT /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, "veiculos:escrever", veiculoPresentation.NewAtualizarHandler(atualizar)))
	mux.Handle("GET /clientes", clientePresentation.NewConsultarHandler(consultarCliente, jwt))
	mux.Handle("POST /clientes", clientePresentation.NewCadastrarHandler(cadastrarCliente, jwt))
	mux.Handle("PUT /clientes/{clienteId}", clientePresentation.NewAtualizarHandler(atualizarCliente, jwt))
	mux.Handle("DELETE /clientes/{clienteId}", clientePresentation.NewInativarHandler(inativarCliente, jwt))
	mux.Handle("POST /clientes/{clienteId}/reativacao", clientePresentation.NewReativarHandler(reativarCliente, jwt))

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
	_, _ = writer.Write([]byte(`{"status":"ok"}`))
}
