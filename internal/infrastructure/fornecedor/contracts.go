package fornecedor

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"

var (
	_ application.AtualizacaoRepository = PostgresRepository{}
	_ application.BuscarPorIDRepository = PostgresRepository{}
	_ application.DesativarRepository   = PostgresRepository{}
	_ application.ListarRepository      = PostgresRepository{}
	_ application.ReativarRepository    = PostgresRepository{}
	_ application.Repository            = PostgresRepository{}
)
