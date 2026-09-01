package notificacao

import (
	"strings"
	"testing"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/notificacao"
)

func orcamentoExemplo() *ResumoOrcamento {
	return &ResumoOrcamento{
		Numero: "PRINCIPAL",
		Itens: []ItemOrcamento{
			{Tipo: "SERVICO", Descricao: "Troca de óleo e filtro", Quantidade: 1, ValorUnitario: 150, ValorTotal: 150},
			{Tipo: "PECA", Descricao: "Filtro de óleo do motor", Quantidade: 2, ValorUnitario: 45, ValorTotal: 90},
			{Tipo: "INSUMO", Descricao: "Óleo sintético 5W30", Quantidade: 3.5, ValorUnitario: 32, ValorTotal: 112},
		},
		ValorTotal:     352,
		EstimativaDias: 3,
	}
}

func TestMensagemOrcamentoTemTabelaHTML(t *testing.T) {
	_, texto, html, err := Mensagem(notificacao.EventoOrcamentoPronto, "Maria", orcamentoExemplo())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "<table") || !strings.Contains(html, "</table>") {
		t.Fatal("o HTML deveria trazer uma tabela")
	}
	for _, esperado := range []string{"Troca de óleo e filtro", "Filtro de óleo do motor", "Óleo sintético 5W30"} {
		if !strings.Contains(html, esperado) {
			t.Fatalf("item ausente no HTML: %q", esperado)
		}
		if !strings.Contains(texto, esperado) {
			t.Fatalf("item ausente no texto: %q", esperado)
		}
	}
	if !strings.Contains(html, "R$ 352,00") || !strings.Contains(texto, "R$ 352,00") {
		t.Fatal("o total deveria aparecer nas duas versões")
	}
	if !strings.Contains(html, "3 dia(s)") {
		t.Fatal("a estimativa deveria aparecer")
	}
}

// A versão em texto precisa continuar legível: nem todo cliente exibe HTML.
func TestMensagemOrcamentoTemVersaoEmTexto(t *testing.T) {
	_, texto, _, err := Mensagem(notificacao.EventoOrcamentoPronto, "Maria", orcamentoExemplo())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(texto, "<table") || strings.Contains(texto, "<td") {
		t.Fatalf("a versão em texto não pode ter marcação HTML: %q", texto)
	}
	if !strings.Contains(texto, "TOTAL") {
		t.Fatal("o texto deveria ter a linha de total")
	}
}

// Sem orçamento o e-mail ainda funciona, só sem tabela.
func TestMensagemOrcamentoSemItens(t *testing.T) {
	for _, resumo := range []*ResumoOrcamento{nil, {}} {
		_, texto, html, err := Mensagem(notificacao.EventoOrcamentoPronto, "Maria", resumo)
		if err != nil {
			t.Fatal(err)
		}
		if html != "" {
			t.Fatal("sem itens não deveria gerar HTML")
		}
		if texto == "" || !strings.Contains(texto, "Maria") {
			t.Fatalf("o texto deveria continuar fazendo sentido: %q", texto)
		}
	}
}

// Os outros eventos continuam só em texto.
func TestOutrosEventosNaoTemHTML(t *testing.T) {
	for _, evento := range []string{notificacao.EventoServicoFinalizado, notificacao.EventoVeiculoEntregue} {
		_, texto, html, err := Mensagem(evento, "Maria", nil)
		if err != nil {
			t.Fatal(err)
		}
		if html != "" {
			t.Fatalf("evento %q não deveria ter HTML", evento)
		}
		if texto == "" {
			t.Fatalf("evento %q ficou sem texto", evento)
		}
	}
}

func TestFormatacaoDeDinheiro(t *testing.T) {
	casos := map[float64]string{
		0: "R$ 0,00", 45: "R$ 45,00", 150.5: "R$ 150,50",
		1234.56: "R$ 1.234,56", 1000000: "R$ 1.000.000,00",
	}
	for valor, esperado := range casos {
		if obtido := dinheiro(valor); obtido != esperado {
			t.Fatalf("dinheiro(%v) = %q, esperado %q", valor, obtido, esperado)
		}
	}
}

// Peça é contada em unidades; insumo pode ser fracionado.
func TestFormatacaoDeQuantidade(t *testing.T) {
	casos := map[float64]string{1: "1", 2: "2", 3.5: "3,500", 0.25: "0,250"}
	for valor, esperado := range casos {
		if obtido := quantidade(valor); obtido != esperado {
			t.Fatalf("quantidade(%v) = %q, esperado %q", valor, obtido, esperado)
		}
	}
}

// Descrição do cliente vai para dentro de HTML: precisa ser escapada.
func TestDescricaoEhEscapadaNoHTML(t *testing.T) {
	resumo := &ResumoOrcamento{
		Itens:      []ItemOrcamento{{Descricao: `<script>alert("x")</script>`, Quantidade: 1, ValorUnitario: 10, ValorTotal: 10}},
		ValorTotal: 10,
	}

	_, _, html, err := Mensagem(notificacao.EventoOrcamentoPronto, "Maria", resumo)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") {
		t.Fatalf("a descrição entrou sem escapar: %q", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Fatal("a descrição deveria aparecer escapada")
	}
}

// Todo evento conhecido precisa ter texto: um evento sem caso no switch cairia no default
// e viraria erro de enfileiramento na hora em que o status mudasse, em produção.
func TestTodoEventoConhecidoTemMensagem(t *testing.T) {
	eventos := []string{
		notificacao.EventoDiagnosticoIniciado,
		notificacao.EventoOrcamentoPronto,
		notificacao.EventoOrcamentoAprovado,
		notificacao.EventoAguardandoRecursos,
		notificacao.EventoRecursosDisponiveis,
		notificacao.EventoExecucaoIniciada,
		notificacao.EventoServicoCancelado,
		notificacao.EventoServicoFinalizado,
		notificacao.EventoVeiculoEntregue,
	}
	for _, evento := range eventos {
		t.Run(evento, func(t *testing.T) {
			if !notificacao.EventoConhecido(evento) {
				t.Fatal("evento fora de EventoConhecido: Nova() vai recusá-lo")
			}
			assunto, conteudo, _, err := Mensagem(evento, "Ana", nil)
			if err != nil {
				t.Fatal(err)
			}
			if assunto == "" || conteudo == "" {
				t.Fatalf("assunto=%q conteudo=%q", assunto, conteudo)
			}
			if !strings.Contains(conteudo, "Ana") {
				t.Fatalf("o corpo deveria tratar o cliente pelo nome: %q", conteudo)
			}
		})
	}
}
