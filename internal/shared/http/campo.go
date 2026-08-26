package http

import "errors"

// ErroCampo liga uma mensagem de validação ao parâmetro que a causou, para que a
// resposta possa preencher o bloco `erros` do problem+json em vez de só descrever a
// falha no `detail`.
type ErroCampo struct {
	Campo    string
	Mensagem string
}

func (erro ErroCampo) Error() string { return erro.Mensagem }

func (erro ErroCampo) CampoInvalido() string { return erro.Campo }

// portadorDeCampo permite que erros de domínio informem o parâmetro culpado sem que os
// pacotes de domínio precisem importar este pacote de HTTP.
type portadorDeCampo interface {
	CampoInvalido() string
}

func NovoErroCampo(campo, mensagem string) ErroCampo {
	return ErroCampo{Campo: campo, Mensagem: mensagem}
}

// CampoDoErro devolve o parâmetro associado ao erro, ou string vazia quando o erro não
// carrega essa informação.
func CampoDoErro(err error) string {
	var portador portadorDeCampo
	if errors.As(err, &portador) {
		return portador.CampoInvalido()
	}
	return ""
}
