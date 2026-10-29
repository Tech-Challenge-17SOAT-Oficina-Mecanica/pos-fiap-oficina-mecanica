package estoque

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"

var (
	_ application.EntradaRepository = PostgresRepository{}
)
