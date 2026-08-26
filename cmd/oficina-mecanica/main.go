package main

import (
	"log"
	"net/http"
	"os"

	clienteApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	fornecedorApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	mecanicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	servicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	veiculoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	fornecedorInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	mecanicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/mecanico"
	pecaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	servicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	clientePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	fornecedorPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	mecanicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/mecanico"
	pecaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	servicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
	veiculoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

const (
	escopoEstoqueLer      = "estoque:ler"
	escopoEstoqueEscrever = "estoque:escrever"
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
	segurancaRepository := segurancaInfrastructure.NewPostgresRepository(db)
	login := segurancaApplication.NewAutenticar(segurancaRepository, jwt)
	mecanicoRepository := mecanicoInfrastructure.NewPostgresRepository(db)
	cadastrarMecanico := mecanicoApplication.NewCadastrar(mecanicoRepository)
	atualizarMecanico := mecanicoApplication.NewAtualizar(mecanicoRepository)
	cadastrar := veiculoApplication.NewCadastrar(veiculo.NewPostgresRepository(db))
	consultar := veiculoApplication.NewConsultar(veiculo.NewPostgresRepository(db))
	atualizar := veiculoApplication.NewAtualizar(veiculo.NewPostgresRepository(db))
	inativar := veiculoApplication.NewInativar(veiculo.NewPostgresRepository(db))
	reativar := veiculoApplication.NewReativar(veiculo.NewPostgresRepository(db))
	pecaRepository := pecaInfrastructure.NewPostgresRepository(db)
	consultarPecas := pecaApplication.NewConsultarPecas(pecaRepository)
	cadastrarPeca := pecaApplication.NewCadastrarPeca(pecaRepository)
	desativarPeca := pecaApplication.NewDesativarPeca(pecaRepository)
	clienteRepository := clienteInfrastructure.NewPostgresRepository(db)
	cadastrarCliente := clienteApplication.NewCadastrar(clienteRepository)
	consultarCliente := clienteApplication.NewConsultar(clienteRepository)
	atualizarCliente := clienteApplication.NewAtualizar(clienteRepository)
	inativarCliente := clienteApplication.NewInativar(clienteRepository)
	reativarCliente := clienteApplication.NewReativar(clienteRepository)
	servicoRepository := servicoInfrastructure.NewPostgresRepository(db)
	cadastrarServico := servicoApplication.NewCadastrar(servicoRepository)
	consultarServico := servicoApplication.NewConsultar(servicoRepository)
	atualizarServico := servicoApplication.NewAtualizar(servicoRepository)
	desativarServico := servicoApplication.NewDesativar(servicoRepository)
	reativarServico := servicoApplication.NewReativar(servicoRepository)
	fornecedorRepository := fornecedorInfrastructure.NewPostgresRepository(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /fornecedores", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewCadastrarHandler(
		fornecedorApplication.NewCadastrar(fornecedorRepository),
	)))
	mux.Handle("GET /fornecedores", segurancaPresentation.RequireScope(jwt, "compras:ler", fornecedorPresentation.NewListarHandler(
		fornecedorApplication.NewConsultarFornecedores(fornecedorRepository),
	)))
	mux.Handle("GET /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:ler", fornecedorPresentation.NewBuscarPorIDHandler(
		fornecedorApplication.NewConsultarFornecedorPorID(fornecedorRepository),
	)))
	mux.Handle("PUT /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewAtualizarHandler(
		fornecedorApplication.NewAtualizarFornecedor(fornecedorRepository),
	)))
	mux.Handle("DELETE /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewDesativarHandler(
		fornecedorApplication.NewDesativarFornecedor(fornecedorRepository),
	)))
	mux.Handle("POST /fornecedores/{fornecedorId}/reativacao", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewReativarHandler(
		fornecedorApplication.NewReativarFornecedor(fornecedorRepository),
	)))
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("POST /mecanicos", segurancaPresentation.RequireScope(jwt, "mecanicos:escrever", mecanicoPresentation.NewCadastrarHandler(cadastrarMecanico)))
	mux.Handle("PUT /mecanicos/{mecanicoId}", segurancaPresentation.RequireScope(jwt, "mecanicos:escrever", mecanicoPresentation.NewAtualizarHandler(atualizarMecanico)))
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
	mux.Handle("POST /estoque/pecas", segurancaPresentation.RequireScope(jwt, escopoEstoqueEscrever,
		pecaPresentation.NewCadastrarPecaHandler(cadastrarPeca)))
	mux.Handle("GET /estoque/pecas", segurancaPresentation.RequireScope(jwt, escopoEstoqueLer,
		pecaPresentation.NewConsultarPecasHandler(consultarPecas)))
	mux.Handle("GET /estoque/pecas/{pecaId}", segurancaPresentation.RequireScope(jwt, escopoEstoqueLer,
		pecaPresentation.NewConsultarPecaPorIDHandler(consultarPecas)))
	mux.Handle("DELETE /estoque/pecas/{pecaId}", segurancaPresentation.RequireScope(jwt, escopoEstoqueEscrever,
		pecaPresentation.NewDesativarPecaHandler(desativarPeca)))
	mux.Handle("POST /servicos", segurancaPresentation.RequireScope(jwt, "servicos:escrever", servicoPresentation.NewCadastrarHandler(cadastrarServico)))
	mux.Handle("GET /servicos", segurancaPresentation.RequireScope(jwt, "servicos:ler", servicoPresentation.NewListarHandler(consultarServico)))
	mux.Handle("GET /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, "servicos:ler", servicoPresentation.NewConsultarHandler(consultarServico)))
	mux.Handle("PATCH /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, "servicos:escrever", servicoPresentation.NewAtualizarHandler(atualizarServico)))
	mux.Handle("DELETE /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, "servicos:escrever", servicoPresentation.NewDesativarHandler(desativarServico)))
	mux.Handle("POST /servicos/{servicoId}/reativacao", segurancaPresentation.RequireScope(jwt, "servicos:escrever", servicoPresentation.NewReativarHandler(reativarServico)))

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
