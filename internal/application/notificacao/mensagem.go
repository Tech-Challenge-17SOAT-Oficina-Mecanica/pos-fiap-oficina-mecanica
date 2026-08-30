package notificacao

import (
	"fmt"
	"html"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

// Mensagem monta o texto de cada evento num lugar so. Quem dispara a notificacao informa
// o que aconteceu, nao como o cliente sera avisado.
//
// Devolve duas versoes do corpo: texto puro, que todo cliente de e-mail entende, e HTML,
// usado quando ha tabela a mostrar. O HTML vazio significa "so texto".
func Mensagem(tipoEvento, nomeCliente string, orcamento *ResumoOrcamento) (assunto, conteudo, conteudoHTML string, err error) {
	tratamento := "Olá"
	if nomeCliente != "" {
		tratamento = fmt.Sprintf("Olá, %s", nomeCliente)
	}

	switch tipoEvento {
	case notificacao.EventoServicoFinalizado:
		return "Seu veículo está pronto para retirada",
			fmt.Sprintf("%s! O serviço do seu veículo foi concluído e ele já está disponível para retirada na oficina. Qualquer dúvida, é só falar com a gente.", tratamento),
			"", nil

	case notificacao.EventoOrcamentoPronto:
		return "Seu orçamento está pronto",
			textoOrcamento(tratamento, orcamento),
			htmlOrcamento(tratamento, orcamento), nil

	case notificacao.EventoVeiculoEntregue:
		return "Confirmação da entrega do seu veículo",
			fmt.Sprintf("%s! Confirmamos a entrega do seu veículo. Obrigado pela confiança, e conte com a gente quando precisar.", tratamento),
			"", nil

	default:
		return "", "", "", notificacao.ErrEventoInvalido
	}
}

const aberturaOrcamento = "! O orçamento do serviço do seu veículo está pronto e aguarda a sua aprovação."

// textoOrcamento monta a versao em texto puro. As colunas usam largura fixa: em fonte
// monoespacada fica alinhado, e em fonte proporcional continua legivel.
func textoOrcamento(tratamento string, orcamento *ResumoOrcamento) string {
	var corpo strings.Builder
	corpo.WriteString(tratamento + aberturaOrcamento + "\n")

	if orcamento == nil || len(orcamento.Itens) == 0 {
		corpo.WriteString("\nAssim que você responder, seguimos com o atendimento.\n")
		return corpo.String()
	}

	if orcamento.Numero != "" {
		fmt.Fprintf(&corpo, "\nOrçamento %s\n", orcamento.Numero)
	}
	corpo.WriteString("\n")
	for _, item := range orcamento.Itens {
		fmt.Fprintf(&corpo, "  %-38s %6s x %10s = %12s\n",
			cortar(item.Descricao, 38), quantidade(item.Quantidade),
			dinheiro(item.ValorUnitario), dinheiro(item.ValorTotal))
	}
	fmt.Fprintf(&corpo, "\n  %-38s %34s\n", "TOTAL", dinheiro(orcamento.ValorTotal))

	if orcamento.EstimativaDias > 0 {
		fmt.Fprintf(&corpo, "\nPrazo estimado de entrega: %d dia(s) após a aprovação.\n", orcamento.EstimativaDias)
	}
	corpo.WriteString("\nAssim que você responder, seguimos com o atendimento.\n")
	return corpo.String()
}

// htmlOrcamento monta a tabela. Os estilos vao embutidos porque cliente de e-mail
// costuma descartar folha de estilo separada.
func htmlOrcamento(tratamento string, orcamento *ResumoOrcamento) string {
	if orcamento == nil || len(orcamento.Itens) == 0 {
		return ""
	}

	var corpo strings.Builder
	corpo.WriteString(`<div style="font-family:Arial,Helvetica,sans-serif;font-size:14px;color:#222;max-width:640px">`)
	fmt.Fprintf(&corpo, `<p>%s%s</p>`, html.EscapeString(tratamento), aberturaOrcamento)

	if orcamento.Numero != "" {
		fmt.Fprintf(&corpo, `<p style="color:#666;margin:4px 0 16px">Orçamento <strong>%s</strong></p>`,
			html.EscapeString(orcamento.Numero))
	}

	corpo.WriteString(`<table style="border-collapse:collapse;width:100%">`)
	corpo.WriteString(`<thead><tr style="background:#f4f4f4">` +
		celulaCabecalho("Item", "left") + celulaCabecalho("Qtd.", "right") +
		celulaCabecalho("Valor unit.", "right") + celulaCabecalho("Total", "right") +
		`</tr></thead><tbody>`)

	for _, item := range orcamento.Itens {
		descricao := html.EscapeString(item.Descricao)
		if item.Tipo != "" {
			descricao += fmt.Sprintf(` <span style="color:#888;font-size:12px">(%s)</span>`,
				html.EscapeString(strings.ToLower(item.Tipo)))
		}
		corpo.WriteString(`<tr>` +
			celula(descricao, "left") + celula(quantidade(item.Quantidade), "right") +
			celula(dinheiro(item.ValorUnitario), "right") + celula(dinheiro(item.ValorTotal), "right") +
			`</tr>`)
	}

	corpo.WriteString(`</tbody><tfoot><tr>` +
		`<td colspan="3" style="padding:10px 8px;text-align:right;border-top:2px solid #ddd"><strong>Total</strong></td>` +
		fmt.Sprintf(`<td style="padding:10px 8px;text-align:right;border-top:2px solid #ddd"><strong>%s</strong></td>`,
			dinheiro(orcamento.ValorTotal)) +
		`</tr></tfoot></table>`)

	if orcamento.EstimativaDias > 0 {
		fmt.Fprintf(&corpo, `<p style="margin-top:16px">Prazo estimado de entrega: <strong>%d dia(s)</strong> após a aprovação.</p>`,
			orcamento.EstimativaDias)
	}
	corpo.WriteString(`<p>Assim que você responder, seguimos com o atendimento.</p></div>`)
	return corpo.String()
}

func celulaCabecalho(texto, alinhamento string) string {
	return fmt.Sprintf(`<th style="padding:8px;text-align:%s;border-bottom:1px solid #ddd;font-size:13px;color:#555">%s</th>`,
		alinhamento, texto)
}

func celula(conteudo, alinhamento string) string {
	return fmt.Sprintf(`<td style="padding:8px;text-align:%s;border-bottom:1px solid #eee">%s</td>`,
		alinhamento, conteudo)
}

// dinheiro formata no padrao brasileiro: 1.234,56
func dinheiro(valor float64) string {
	texto := fmt.Sprintf("%.2f", valor)
	inteiro, decimal, _ := strings.Cut(texto, ".")

	var comMilhar strings.Builder
	for indice, digito := range inteiro {
		if indice > 0 && (len(inteiro)-indice)%3 == 0 {
			comMilhar.WriteByte('.')
		}
		comMilhar.WriteRune(digito)
	}
	return "R$ " + comMilhar.String() + "," + decimal
}

// quantidade omite as casas decimais quando o numero e inteiro: peca e contada em
// unidades, insumo pode ser fracionado.
func quantidade(valor float64) string {
	if valor == float64(int64(valor)) {
		return fmt.Sprintf("%d", int64(valor))
	}
	return strings.Replace(fmt.Sprintf("%.3f", valor), ".", ",", 1)
}

func cortar(texto string, limite int) string {
	if len([]rune(texto)) <= limite {
		return texto
	}
	return string([]rune(texto)[:limite-1]) + "…"
}
