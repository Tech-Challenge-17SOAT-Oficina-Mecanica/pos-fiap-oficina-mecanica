package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	clienteApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	estoqueApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	fornecedorApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	insumoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	mecanicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	notificacaoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"
	orcamentoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	ordemServicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	servicoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	veiculoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	segurancaDominio "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
	clienteInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/cliente"
	estoqueInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/estoque"
	fornecedorInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	insumoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/insumo"
	mecanicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/mecanico"
	notificacaoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/notificacao"
	orcamentoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/orcamento"
	ordemServicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
	pecaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	servicoInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/servico"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	clientePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/cliente"
	estoquePresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/estoque"
	fornecedorPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	insumoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/insumo"
	mecanicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/mecanico"
	orcamentoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/orcamento"
	ordemServicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/ordemservico"
	pecaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	servicoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/servico"
	veiculoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
	sharedhttp "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/http"
)

func montarAplicacao(db *pgxpool.Pool, jwt segurancaInfrastructure.JWT, life *serverLifecycle) (http.Handler, notificacaoApplication.Processar) {
	notificacaoRepository := notificacaoInfrastructure.NewPostgresRepository(db)
	enfileirarNotificacao := notificacaoApplication.NewEnfileirar(notificacaoRepository)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ready", life.readiness(readyHandler(db)))
	registrarRotasSeguranca(mux, db, jwt)
	registrarRotasMecanicos(mux, db, jwt)
	registrarRotasClientes(mux, db, jwt)
	registrarRotasVeiculos(mux, db, jwt)
	registrarRotasFornecedores(mux, db, jwt)
	registrarRotasServicos(mux, db, jwt)
	registrarRotasOrdensServico(mux, db, jwt, enfileirarNotificacao)
	registrarRotasOrcamentos(mux, db, jwt, enfileirarNotificacao)
	registrarRotasEstoque(mux, db, jwt, enfileirarNotificacao)
	return sharedhttp.CORS(mux), notificacaoApplication.NewProcessar(notificacaoRepository, enviadorDeNotificacao())
}

func registrarRotasSeguranca(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT) {
	segurancaRepository := segurancaInfrastructure.NewPostgresRepository(db)
	login := segurancaApplication.NewAutenticar(segurancaRepository, jwt)

	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
}

func registrarRotasMecanicos(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT) {
	mecanicoRepository := mecanicoInfrastructure.NewPostgresRepository(db)
	cadastrarMecanico := mecanicoApplication.NewCadastrar(mecanicoRepository)
	atualizarMecanico := mecanicoApplication.NewAtualizar(mecanicoRepository)

	mux.Handle("POST /mecanicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoMecanicosEscrever, mecanicoPresentation.NewCadastrarHandler(cadastrarMecanico)))
	mux.Handle("PUT /mecanicos/{mecanicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoMecanicosEscrever, mecanicoPresentation.NewAtualizarHandler(atualizarMecanico)))
}

func registrarRotasClientes(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT) {
	clienteRepository := clienteInfrastructure.NewPostgresRepository(db)
	cadastrarCliente := clienteApplication.NewCadastrar(clienteRepository)
	consultarCliente := clienteApplication.NewConsultar(clienteRepository)
	atualizarCliente := clienteApplication.NewAtualizar(clienteRepository)
	inativarCliente := clienteApplication.NewInativar(clienteRepository)
	reativarCliente := clienteApplication.NewReativar(clienteRepository)

	mux.Handle("GET /clientes", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesLer, clientePresentation.NewConsultarHandler(consultarCliente)))
	mux.Handle("POST /clientes", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewCadastrarHandler(cadastrarCliente)))
	mux.Handle("PUT /clientes/{clienteId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewAtualizarHandler(atualizarCliente)))
	mux.Handle("DELETE /clientes/{clienteId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewInativarHandler(inativarCliente)))
	mux.Handle("POST /clientes/{clienteId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoClientesEscrever, clientePresentation.NewReativarHandler(reativarCliente)))
}

func registrarRotasVeiculos(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT) {
	veiculoRepository := veiculo.NewPostgresRepository(db)
	cadastrarVeiculo := veiculoApplication.NewCadastrar(veiculoRepository)
	consultarVeiculo := veiculoApplication.NewConsultar(veiculoRepository)
	atualizarVeiculo := veiculoApplication.NewAtualizar(veiculoRepository)
	inativarVeiculo := veiculoApplication.NewInativar(veiculoRepository)
	reativarVeiculo := veiculoApplication.NewReativar(veiculoRepository)

	mux.Handle("POST /clientes/{clienteId}/veiculos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewCadastrarHandler(cadastrarVeiculo)))
	mux.Handle("GET /veiculos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosLer, veiculoPresentation.NewConsultaHandler(consultarVeiculo)))
	mux.Handle("PUT /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewAtualizarHandler(atualizarVeiculo)))
	mux.Handle("DELETE /veiculos/{veiculoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewInativarHandler(inativarVeiculo)))
	mux.Handle("POST /veiculos/{veiculoId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoVeiculosEscrever, veiculoPresentation.NewReativarHandler(reativarVeiculo)))
}

func registrarRotasFornecedores(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT) {
	fornecedorRepository := fornecedorInfrastructure.NewPostgresRepository(db)

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
}

func registrarRotasServicos(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT) {
	servicoRepository := servicoInfrastructure.NewPostgresRepository(db)
	cadastrarServico := servicoApplication.NewCadastrar(servicoRepository)
	consultarServico := servicoApplication.NewConsultar(servicoRepository)
	atualizarServico := servicoApplication.NewAtualizar(servicoRepository)
	desativarServico := servicoApplication.NewDesativar(servicoRepository)
	reativarServico := servicoApplication.NewReativar(servicoRepository)

	mux.Handle("POST /servicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewCadastrarHandler(cadastrarServico)))
	mux.Handle("GET /servicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosLer, servicoPresentation.NewListarHandler(consultarServico)))
	mux.Handle("GET /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosLer, servicoPresentation.NewConsultarHandler(consultarServico)))
	mux.Handle("PATCH /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewAtualizarHandler(atualizarServico)))
	mux.Handle("DELETE /servicos/{servicoId}", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewDesativarHandler(desativarServico)))
	mux.Handle("POST /servicos/{servicoId}/reativacao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoServicosEscrever, servicoPresentation.NewReativarHandler(reativarServico)))
}

func registrarRotasOrdensServico(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT, enfileirarNotificacao notificacaoApplication.Enfileirar) {
	ordemServicoRepository := ordemServicoInfrastructure.NewPostgresRepository(db)
	criarOrdemServico := ordemServicoApplication.NewCriar(ordemServicoRepository)
	registrarProblemaRelatado := ordemServicoApplication.NewRegistrarProblemaRelatado(ordemServicoRepository, enfileirarNotificacao, nil)
	registrarProblema := ordemServicoApplication.NewRegistrarProblema(ordemServicoRepository)
	registrarServicos := ordemServicoApplication.NewRegistrarServicos(ordemServicoRepository)
	registrarItensOS := ordemServicoApplication.NewRegistrarItens(ordemServicoRepository)
	iniciarExecucao := ordemServicoApplication.NewIniciarExecucao(ordemServicoRepository, enfileirarNotificacao, nil)
	finalizarServico := ordemServicoApplication.NewFinalizar(ordemServicoRepository, enfileirarNotificacao, nil)
	registrarEntrega := ordemServicoApplication.NewEntregar(ordemServicoRepository, enfileirarNotificacao, nil)
	consultarOS := ordemServicoApplication.NewConsultar(ordemServicoRepository)
	listarOS := ordemServicoApplication.NewListar(ordemServicoRepository)
	consultarTempoExecucao := ordemServicoApplication.NewConsultarTempoExecucaoDaOS(ordemServicoRepository)
	consultarTempoMedioExecucao := ordemServicoApplication.NewConsultarTempoMedioExecucao(ordemServicoRepository)
	consultarFila := ordemServicoApplication.NewConsultarFila(ordemServicoRepository)

	mux.Handle("POST /ordens-servico/{osId}/pecas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarPecasHandler(registrarItensOS)))
	mux.Handle("POST /ordens-servico/{osId}/insumos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarInsumosHandler(registrarItensOS)))
	mux.Handle("POST /ordens-servico/{osId}/problemas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarProblemaHandler(registrarProblema)))
	mux.Handle("POST /ordens-servico", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewCriarHandler(criarOrdemServico)))
	mux.Handle("POST /ordens-servico/{osId}/problema-relatado", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarProblemaRelatadoHandler(registrarProblemaRelatado)))
	mux.Handle("POST /ordens-servico/{osId}/servicos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewRegistrarServicosHandler(registrarServicos)))
	mux.Handle("POST /ordens-servico/{osId}/execucao/iniciar", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewIniciarExecucaoHandler(iniciarExecucao)))
	mux.Handle("POST /ordens-servico/{osId}/finalizar", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewFinalizarHandler(finalizarServico)))
	mux.Handle("GET /ordens-servico/{osId}/tempo-execucao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSLer, ordemServicoPresentation.NewConsultarTempoExecucaoHandler(consultarTempoExecucao)))
	mux.Handle("GET /ordens-servico/tempos-execucao", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSLer, ordemServicoPresentation.NewListarTemposExecucaoHandler(consultarTempoMedioExecucao)))
	mux.Handle("POST /ordens-servico/{osId}/entrega", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSEscrever, ordemServicoPresentation.NewEntregarHandler(registrarEntrega)))
	mux.Handle("GET /ordens-servico/{osId}", segurancaPresentation.RequireAnyScope(jwt, []string{segurancaDominio.EscopoOSLer, segurancaDominio.EscopoOrcamentosLer}, ordemServicoPresentation.NewConsultarHandler(consultarOS)))
	mux.Handle("GET /ordens-servico", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSLer, ordemServicoPresentation.NewListarHandler(listarOS)))
	mux.Handle("GET /fila-atendimento", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOSLer, ordemServicoPresentation.NewConsultarFilaHandler(consultarFila)))
}

func registrarRotasOrcamentos(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT, enfileirarNotificacao notificacaoApplication.Enfileirar) {
	// Zero deixa a fila fora da estimativa de entrega.
	capacidadeDiariaOS := inteiroDoAmbiente("CAPACIDADE_DIARIA_OS", 0)
	orcamentoRepository := orcamentoInfrastructure.NewPostgresRepository(db)
	consultarOrcamento := orcamentoApplication.NewConsultar(orcamentoRepository)
	aprovarOrcamento := orcamentoApplication.NewAprovar(orcamentoRepository, enfileirarNotificacao, nil)
	recusarOrcamento := orcamentoApplication.NewRecusar(orcamentoRepository, enfileirarNotificacao, nil)

	mux.Handle("POST /orcamentos/{orcamentoId}/enviar", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOrcamentosEscrever,
		orcamentoPresentation.NewEnviarHandler(orcamentoApplication.NewEnviar(orcamentoRepository, enfileirarNotificacao, nil))))
	mux.Handle("POST /orcamentos/{orcamentoId}/calcular", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOrcamentosEscrever,
		orcamentoPresentation.NewCalcularHandler(orcamentoApplication.NewCalcular(orcamentoRepository, capacidadeDiariaOS))))
	mux.Handle("GET /ordens-servico/{osId}/orcamento", segurancaPresentation.RequireAnyScope(jwt, []string{segurancaDominio.EscopoOSLer, segurancaDominio.EscopoOrcamentosLer}, orcamentoPresentation.NewConsultarHandler(consultarOrcamento)))
	mux.Handle("POST /orcamentos/{orcamentoId}/aprovar", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOrcamentosDecidir, orcamentoPresentation.NewAprovarHandler(aprovarOrcamento)))
	mux.Handle("POST /orcamentos/{orcamentoId}/recusar", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoOrcamentosDecidir, orcamentoPresentation.NewRecusarHandler(recusarOrcamento)))
}

func registrarRotasEstoque(mux *http.ServeMux, db *pgxpool.Pool, jwt segurancaInfrastructure.JWT, enfileirarNotificacao notificacaoApplication.Enfileirar) {
	pecaRepository := pecaInfrastructure.NewPostgresRepository(db)
	consultarPecas := pecaApplication.NewConsultarPecas(pecaRepository)
	cadastrarPeca := pecaApplication.NewCadastrarPeca(pecaRepository)
	processarPecas := pecaApplication.NewSolicitarCompraEReservarPecas(pecaRepository)
	atualizarPeca := pecaApplication.NewAtualizarPeca(pecaRepository)
	insumoRepository := insumoInfrastructure.NewPostgresRepository(db)
	cadastrarInsumo := insumoApplication.NewCadastrarInsumo(insumoRepository)
	consultarInsumos := insumoApplication.NewConsultarInsumos(insumoRepository)
	desativarInsumo := insumoApplication.NewDesativarInsumo(insumoRepository)
	processarInsumos := insumoApplication.NewSolicitarCompraEReservarInsumos(insumoRepository)
	desativarPeca := pecaApplication.NewDesativarPeca(pecaRepository)
	estoqueRepository := estoqueInfrastructure.NewPostgresRepository(db)
	registrarEntrada := estoqueApplication.NewRegistrarEntrada(estoqueRepository, enfileirarNotificacao, nil)

	mux.Handle("POST /estoque/pecas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
		pecaPresentation.NewCadastrarPecaHandler(cadastrarPeca)))
	mux.Handle("POST /estoque/solicitacoes-compra-reserva", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueMovimentar,
		pecaPresentation.NewSolicitarCompraEReservarPecasHandler(processarPecas)))
	mux.Handle("POST /estoque/insumos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueEscrever,
		insumoPresentation.NewCadastrarInsumoHandler(cadastrarInsumo)))
	mux.Handle("POST /estoque/solicitacoes-compra-reserva-insumos", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueMovimentar,
		insumoPresentation.NewSolicitarCompraEReservarInsumosHandler(processarInsumos)))
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
	mux.Handle("POST /estoque/entradas", segurancaPresentation.RequireScope(jwt, segurancaDominio.EscopoEstoqueMovimentar,
		estoquePresentation.NewRegistrarEntradaHandler(registrarEntrada)))
}
