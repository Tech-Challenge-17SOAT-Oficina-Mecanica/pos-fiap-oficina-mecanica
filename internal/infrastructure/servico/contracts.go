package servico

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"

var (
	_ application.AtualizarRepository = PostgresRepository{}
	_ application.ConsultarRepository = PostgresRepository{}
	_ application.DesativarRepository = PostgresRepository{}
	_ application.ReativarRepository  = PostgresRepository{}
	_ application.Repository          = PostgresRepository{}
)
