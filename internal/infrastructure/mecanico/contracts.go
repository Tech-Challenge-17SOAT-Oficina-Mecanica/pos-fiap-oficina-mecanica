package mecanico

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"

var (
	_ application.AtualizarRepository = PostgresRepository{}
	_ application.CadastrarRepository = PostgresRepository{}
)
