package ordemservico

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"

var (
	_ application.ConsultarFilaRepository          = PostgresRepository{}
	_ application.ConsultarRepository              = PostgresRepository{}
	_ application.ConsultarTempoExecucaoRepository = PostgresRepository{}
	_ application.DevolucaoRepository              = PostgresRepository{}
	_ application.EntregarRepository               = PostgresRepository{}
	_ application.FinalizarRepository              = PostgresRepository{}
	_ application.IniciarExecucaoRepository        = PostgresRepository{}
	_ application.ItensRepository                  = PostgresRepository{}
	_ application.ListarRepository                 = PostgresRepository{}
	_ application.ListarTemposExecucaoRepository   = PostgresRepository{}
	_ application.ProblemaRelatadoRepository       = PostgresRepository{}
	_ application.ProblemaRepository               = PostgresRepository{}
	_ application.Repository                       = PostgresRepository{}
	_ application.ServicoRepository                = PostgresRepository{}
)
