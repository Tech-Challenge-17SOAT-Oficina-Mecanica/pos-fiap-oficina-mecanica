package notificacao

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/notificacao"

var (
	_ application.FilaRepository = PostgresRepository{}
	_ application.Repository     = PostgresRepository{}
)
