package cliente

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"

var (
	_ application.AtualizarRepository = PostgresRepository{}
	_ application.CadastrarRepository = PostgresRepository{}
	_ application.ConsultarRepository = PostgresRepository{}
	_ application.InativarRepository  = PostgresRepository{}
	_ application.ReativarRepository  = PostgresRepository{}
)
