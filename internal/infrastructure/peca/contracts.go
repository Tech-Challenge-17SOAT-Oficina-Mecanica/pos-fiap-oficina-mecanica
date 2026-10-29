package peca

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"

var (
	_ application.AtualizarRepository      = PostgresRepository{}
	_ application.CadastrarRepository      = PostgresRepository{}
	_ application.DesativarRepository      = PostgresRepository{}
	_ application.ProcessarPecasRepository = PostgresRepository{}
	_ application.Repository               = PostgresRepository{}
)
