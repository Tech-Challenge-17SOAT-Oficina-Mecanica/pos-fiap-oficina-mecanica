package notificacao

import (
	"context"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// Enviador e a porta de saida: a unica coisa que sabe o que e um canal de verdade.
// Trocar log por SMTP e implementar esta interface de novo, sem tocar no resto.
type Enviador interface {
	Enviar(ctx context.Context, aviso notificacao.Notificacao) error
}

type FilaRepository interface {
	// ProcessarPendentes reserva cada mensagem ate gravar seu resultado.
	// A funcao de tratamento executa com exclusividade entre consumidores.
	ProcessarPendentes(context.Context, int, func(context.Context, notificacao.Notificacao) (notificacao.Notificacao, error)) error
}

// Processar consome a fila. E aqui que o envio acontece — fora da transacao de quem
// disparou, para que a falha nunca desfaca a operacao de negocio.
type Processar struct {
	repository FilaRepository
	enviador   Enviador
	agora      func() time.Time
}

func NewProcessar(repository FilaRepository, enviador Enviador) Processar {
	return Processar{repository: repository, enviador: enviador, agora: time.Now}
}

type Resultado struct {
	Processadas int
	Enviadas    int
	Falhas      int
}

// Execute tenta enviar ate `limite` notificacoes. Uma falha individual nao interrompe as
// demais: a notificacao fica marcada como FALHOU e volta na proxima rodada.
func (useCase Processar) Execute(ctx context.Context, limite int) (Resultado, error) {
	var resultado Resultado
	err := useCase.repository.ProcessarPendentes(ctx, limite, func(ctx context.Context, aviso notificacao.Notificacao) (notificacao.Notificacao, error) {
		if err := ctx.Err(); err != nil {
			return aviso, err
		}
		envioCtx, cancelar := context.WithTimeout(ctx, 10*time.Second)
		err := useCase.enviador.Enviar(envioCtx, aviso)
		cancelar()
		if ctx.Err() != nil {
			return aviso, ctx.Err()
		}
		resultado.Processadas++
		if err != nil {
			resultado.Falhas++
			return aviso.MarcarFalha(err.Error()), nil
		}
		enviada, err := aviso.MarcarEnviada(useCase.agora())
		if err != nil {
			return aviso, err
		}
		resultado.Enviadas++
		return enviada, nil
	})
	return resultado, err
}
