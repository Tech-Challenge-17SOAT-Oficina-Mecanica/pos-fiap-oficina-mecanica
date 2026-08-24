package seguranca

type Usuario struct {
	ID        string
	Email     string
	SenhaHash string
	Ativo     bool
	Escopos   []string
}
