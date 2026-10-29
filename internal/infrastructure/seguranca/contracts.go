package seguranca

import application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"

var (
	_ application.UsuarioRepository = PostgresRepository{}
)
