package notificacao

import (
	"fmt"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// Mensagem monta o texto de cada evento num lugar so. Quem dispara a notificacao informa
// o que aconteceu, nao como o cliente sera avisado.
func Mensagem(tipoEvento, nomeCliente string) (assunto string, conteudo string, err error) {
	tratamento := "Olá"
	if nomeCliente != "" {
		tratamento = fmt.Sprintf("Olá, %s", nomeCliente)
	}

	switch tipoEvento {
	case notificacao.EventoServicoFinalizado:
		return "Seu veículo está pronto para retirada",
			fmt.Sprintf("%s! O serviço do seu veículo foi concluído e ele já está disponível para retirada na oficina. Qualquer dúvida, é só falar com a gente.", tratamento), nil

	case notificacao.EventoOrcamentoPronto:
		return "Seu orçamento está pronto",
			fmt.Sprintf("%s! O orçamento do serviço do seu veículo está pronto e aguarda a sua aprovação. Assim que você responder, seguimos com o atendimento.", tratamento), nil

	case notificacao.EventoVeiculoEntregue:
		return "Confirmação da entrega do seu veículo",
			fmt.Sprintf("%s! Confirmamos a entrega do seu veículo. Obrigado pela confiança, e conte com a gente quando precisar.", tratamento), nil

	default:
		return "", "", notificacao.ErrEventoInvalido
	}
}
