package main

import (
	"log"
	"net/http"
	"os"

	clienteApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	fornecedorApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	insumoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	mecanicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	orcamentoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	ordemServicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	servicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	veiculoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	segurancaDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	fornecedorInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	insumoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/insumo"
	mecanicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/mecanico"
	orcamentoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	ordemServicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	pecaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	servicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	clientePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	fornecedorPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	insumoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/insumo"
	mecanicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/mecanico"
	orcamentoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	ordemServicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	pecaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	servicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
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
	atualizarPeca := pecaApplication.NewAtualizarPeca(pecaRepository)
	insumoRepository := insumoInfrastructure.NewPostgresRepository(db)
	cadastrarInsumo := insumoApplication.NewCadastrarInsumo(insumoRepository)
	consultarInsumos := insumoApplication.NewConsultarInsumos(insumoRepository)
	desativarInsumo := insumoApplication.NewDesativarInsumo(insumoRepository)
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
	ordemServicoRepository := ordemServicoInfrastructure.NewPostgresRepository(db)
	criarOrdemServico := ordemServicoApplication.NewCriar(ordemServicoRepository)
	registrarProblemaRelatado := ordemServicoApplication.NewRegistrarProblemaRelatado(ordemServicoRepository)
	registrarProblema := ordemServicoApplication.NewRegistrarProblema(ordemServicoRepository)
	registrarServicos := ordemServicoApplication.NewRegistrarServicos(ordemServicoRepository)
	orcamentoRepository := orcamentoInfrastructure.NewPostgresRepository(db)
	consultarOrcamento := orcamentoApplication.NewConsultar(orcamentoRepository)
	fornecedorRepository := fornecedorInfrastructure.NewPostgresRepository(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /fornecedores", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoComprasEscrever, fornecedorPresentation.NewCadastrarHandler(
		fornecedorApplication.NewCadastrar(fornecedorRepository),
	)))
	mux.Handle("GET /fornecedores", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoComprasLer, fornecedorPresentation.NewListarHandler(
		fornecedorApplication.NewConsultarFornecedores(fornecedorRepository),
	)))
	mux.Handle("GET /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoComprasLer, fornecedorPresentation.NewBuscarPorIDHandler(
		fornecedorApplication.NewConsultarFornecedorPorID(fornecedorRepository),
	)))
	mux.Handle("PUT /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoComprasEscrever, fornecedorPresentation.NewAtualizarHandler(
		fornecedorApplication.NewAtualizarFornecedor(fornecedorRepository),
	)))
	mux.Handle("DELETE /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoComprasEscrever, fornecedorPresentation.NewDesativarHandler(
		fornecedorApplication.NewDesativarFornecedor(fornecedorRepository),
	)))
	mux.Handle("POST /fornecedores/{fornecedorId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoComprasEscrever, fornecedorPresentation.NewReativarHandler(
		fornecedorApplication.NewReativarFornecedor(fornecedorRepository),
	)))
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("POST /mecanicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoMecanicosEscrever, mecanicoPresentation.NewCadastrarHandler(cadastrarMecanico)))
	mux.Handle("PUT /mecanicos/{mecanicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoMecanicosEscrever, mecanicoPresentation.NewAtualizarHandler(atualizarMecanico)))
	mux.Handle("POST /clientes/{clienteId}/veiculos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewHandler(cadastrar)))
	mux.Handle("POST /ordens-servico/{osId}/problemas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarProblemaHandler(registrarProblema)))
	mux.Handle("POST /ordens-servico", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewCriarHandler(criarOrdemServico)))
	mux.Handle("POST /ordens-servico/{osId}/problema-relatado", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarProblemaRelatadoHandler(registrarProblemaRelatado)))
	mux.Handle("POST /ordens-servico/{osId}/servicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarServicosHandler(registrarServicos)))
	mux.Handle("POST /orcamentos/{orcamentoId}/calcular", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOrcamentosEscrever,
		orcamentoPresentation.NewCalcularHandler(orcamentoApplication.NewCalcular(orcamentoRepository))))
	mux.Handle("GET /ordens-servico/{osId}/orcamento", segurancaPresentation.RequireAnyScope(jwt, []string{segurancaDominio.EscopoOSLer, segurancaDominio.EscopoOrcamentosLer}, orcamentoPresentation.NewConsultarHandler(consultarOrcamento)))
	mux.Handle("GET /veiculos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosLer, veiculoPresentation.NewConsultaHandler(consultar)))
	mux.Handle("PUT /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewAtualizarHandler(atualizar)))
	mux.Handle("DELETE /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewInativarHandler(inativar)))
	mux.Handle("POST /veiculos/{veiculoId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewReativarHandler(reativar)))
	mux.Handle("GET /clientes", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesLer, clientePresentation.NewConsultarHandler(consultarCliente)))
	mux.Handle("POST /clientes", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewCadastrarHandler(cadastrarCliente)))
	mux.Handle("PUT /clientes/{clienteId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewAtualizarHandler(atualizarCliente)))
	mux.Handle("DELETE /clientes/{clienteId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewInativarHandler(inativarCliente)))
	mux.Handle("POST /clientes/{clienteId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewReativarHandler(reativarCliente)))
	mux.Handle("POST /servicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewCadastrarHandler(cadastrarServico)))
	mux.Handle("GET /servicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosLer, servicoPresentation.NewListarHandler(consultarServico)))
	mux.Handle("GET /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosLer, servicoPresentation.NewConsultarHandler(consultarServico)))
	mux.Handle("PATCH /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewAtualizarHandler(atualizarServico)))
	mux.Handle("DELETE /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewDesativarHandler(desativarServico)))
	mux.Handle("POST /servicos/{servicoId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewReativarHandler(reativarServico)))
	mux.Handle("POST /estoque/pecas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
		pecaPresentation.NewCadastrarPecaHandler(cadastrarPeca)))
	mux.Handle("POST /estoque/insumos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
		insumoPresentation.NewCadastrarInsumoHandler(cadastrarInsumo)))
	mux.Handle("GET /estoque/insumos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueLer,
		insumoPresentation.NewConsultarInsumosHandler(consultarInsumos)))
	mux.Handle("GET /estoque/insumos/{insumoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueLer,
		insumoPresentation.NewConsultarInsumoPorIDHandler(consultarInsumos)))
	mux.Handle("DELETE /estoque/insumos/{insumoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
		insumoPresentation.NewDesativarHandler(desativarInsumo)))
	mux.Handle("GET /estoque/pecas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueLer,
		pecaPresentation.NewConsultarPecasHandler(consultarPecas)))
	mux.Handle("GET /estoque/pecas/{pecaId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueLer,
		pecaPresentation.NewConsultarPecaPorIDHandler(consultarPecas)))
	mux.Handle("PUT /estoque/pecas/{pecaId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
		pecaPresentation.NewAtualizarPecaHandler(atualizarPeca)))
	mux.Handle("DELETE /estoque/pecas/{pecaId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
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
	_, _ = writer.Write([]byte(`{"status":"ok"}`))
}
