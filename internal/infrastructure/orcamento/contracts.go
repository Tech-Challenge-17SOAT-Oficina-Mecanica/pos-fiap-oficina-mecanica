package orcamento

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"

var (
	_ application.AprovarRepository  = PostgresRepository{}
	_ application.CalcularRepository = PostgresRepository{}
	_ application.EnviarRepository   = PostgresRepository{}
	_ application.RecusarRepository  = PostgresRepository{}
	_ application.Repository         = PostgresRepository{}
)
