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
	// Pendentes devolve as notificacoes ainda nao entregues, das mais antigas primeiro.
	Pendentes(ctx context.Context, limite int) ([]notificacao.Notificacao, error)
	// AtualizarResultado grava o desfecho da tentativa.
	AtualizarResultado(ctx context.Context, aviso notificacao.Notificacao) error
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
	pendentes, err := useCase.repository.Pendentes(ctx, limite)
	if err != nil {
		return Resultado{}, err
	}

	var resultado Resultado
	for _, aviso := range pendentes {
		resultado.Processadas++

		if err := useCase.enviador.Enviar(ctx, aviso); err != nil {
			if erroAoGravar := useCase.repository.AtualizarResultado(ctx, aviso.MarcarFalha(err.Error())); erroAoGravar != nil {
				return resultado, erroAoGravar
			}
			resultado.Falhas++
			continue
		}

		enviada, err := aviso.MarcarEnviada(useCase.agora())
		if err != nil {
			return resultado, err
		}
		if err := useCase.repository.AtualizarResultado(ctx, enviada); err != nil {
			return resultado, err
		}
		resultado.Enviadas++
	}
	return resultado, nil
}
