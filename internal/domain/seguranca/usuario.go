package seguranca

type Usuario struct {
	ID        string
	Email     string
	SenhaHash string
	Ativo     bool
	Escopos   []string
}

type Claims struct {
	UsuarioID      string
	ClienteID      string
	OrdemServicoID string
	Escopos        []string
}
