package main

import (
	"log"
	"net/http"
	"os"

	clienteApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	ordemServicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	veiculoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	ordemServicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	clientePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	ordemServicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	veiculoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

func main() {
	db, err := database.OpenPool()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	jwt, err := segurancaInfrastructure.NewJWT(os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	login := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt)
	cadastrar := veiculoApplication.NewCadastrar(veiculo.NewPostgresRepository(db))
	consultar := veiculoApplication.NewConsultar(veiculo.NewPostgresRepository(db))
	atualizar := veiculoApplication.NewAtualizar(veiculo.NewPostgresRepository(db))
	inativar := veiculoApplication.NewInativar(veiculo.NewPostgresRepository(db))
	reativar := veiculoApplication.NewReativar(veiculo.NewPostgresRepository(db))
	clienteRepository := clienteInfrastructure.NewPostgresRepository(db)
	cadastrarCliente := clienteApplication.NewCadastrar(clienteRepository)
	consultarCliente := clienteApplication.NewConsultar(clienteRepository)
	atualizarCliente := clienteApplication.NewAtualizar(clienteRepository)
	inativarCliente := clienteApplication.NewInativar(clienteRepository)
	reativarCliente := clienteApplication.NewReativar(clienteRepository)
	registrarProblemaRelatado := ordemServicoApplication.NewRegistrarProblemaRelatado(ordemServicoInfrastructure.NewPostgresRepository(db))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("POST /clientes/{clienteId}/veiculos", segurancaPresentation.RequireScope(jwt, "veiculos:escrever", veiculoPresentation.NewHandler(cadastrar)))
	mux.Handle("GET /veiculos", segurancaPresentation.RequireScope(jwt, "veiculos:ler", veiculoPresentation.NewConsultaHandler(consultar)))
	mux.Handle("PUT /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, "veiculos:escrever", veiculoPresentation.NewAtualizarHandler(atualizar)))
	mux.Handle("DELETE /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, "veiculos:escrever", veiculoPresentation.NewInativarHandler(inativar)))
	mux.Handle("POST /veiculos/{veiculoId}/reativacao", segurancaPresentation.RequireScope(jwt, "veiculos:escrever", veiculoPresentation.NewReativarHandler(reativar)))
	mux.Handle("GET /clientes", segurancaPresentation.RequireScope(jwt, "clientes:ler", clientePresentation.NewConsultarHandler(consultarCliente)))
	mux.Handle("POST /clientes", segurancaPresentation.RequireScope(jwt, "clientes:escrever", clientePresentation.NewCadastrarHandler(cadastrarCliente)))
	mux.Handle("PUT /clientes/{clienteId}", segurancaPresentation.RequireScope(jwt, "clientes:escrever", clientePresentation.NewAtualizarHandler(atualizarCliente)))
	mux.Handle("DELETE /clientes/{clienteId}", segurancaPresentation.RequireScope(jwt, "clientes:escrever", clientePresentation.NewInativarHandler(inativarCliente)))
	mux.Handle("POST /clientes/{clienteId}/reativacao", segurancaPresentation.RequireScope(jwt, "clientes:escrever", clientePresentation.NewReativarHandler(reativarCliente)))
	mux.Handle("POST /ordens-servico/{osId}/problema-relatado", segurancaPresentation.RequireScope(jwt, "os:escrever", ordemServicoPresentation.NewRegistrarProblemaRelatadoHandler(registrarProblemaRelatado)))

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
