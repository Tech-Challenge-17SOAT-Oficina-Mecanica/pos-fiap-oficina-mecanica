package insumo

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"

var (
	_ application.CadastrarRepository        = PostgresRepository{}
	_ application.ConsultarRepository        = PostgresRepository{}
	_ application.DesativarRepository        = PostgresRepository{}
	_ application.ProcessarInsumosRepository = PostgresRepository{}
)
