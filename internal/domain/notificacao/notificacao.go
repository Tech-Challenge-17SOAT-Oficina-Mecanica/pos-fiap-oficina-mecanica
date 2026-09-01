// Package notificacao concentra o aviso ao cliente. O dominio nao sabe o que e e-mail:
// descreve o que precisa ser comunicado e deixa o canal para a infraestrutura.
package notificacao

import (
	"errors"
	"strings"
	"time"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/validation"
)

const (
	CanalEmail = "EMAIL"

	StatusPendente = "PENDENTE"
	StatusEnviada  = "ENVIADA"
	StatusFalhou   = "FALHOU"
)

// Tipos de evento que geram aviso. Novos gatilhos entram aqui, nao em cada modulo.
//
// A lista acompanha a maquina de estados da Ordem de Servico: cada mudanca de status
// visivel para o cliente tem o seu evento. AGUARDANDO_EXECUCAO aparece em dois deles
// porque a OS chega la por dois caminhos diferentes, e o que o cliente precisa saber
// muda: aprovou o orcamento, ou as pecas que faltavam chegaram.
const (
	EventoDiagnosticoIniciado = "DIAGNOSTICO_INICIADO" // OS -> EM_DIAGNOSTICO
	EventoOrcamentoPronto     = "ORCAMENTO_PRONTO"     // OS -> AGUARDANDO_APROVACAO
	EventoOrcamentoAprovado   = "ORCAMENTO_APROVADO"   // OS -> AGUARDANDO_EXECUCAO
	EventoAguardandoRecursos  = "AGUARDANDO_RECURSOS"  // OS -> AGUARDANDO_RECURSOS
	EventoRecursosDisponiveis = "RECURSOS_DISPONIVEIS" // OS -> AGUARDANDO_EXECUCAO
	EventoExecucaoIniciada    = "EXECUCAO_INICIADA"    // OS -> EM_EXECUCAO
	EventoServicoCancelado    = "SERVICO_CANCELADO"    // OS -> CANCELADA
	EventoServicoFinalizado   = "SERVICO_FINALIZADO"   // OS -> FINALIZADA
	EventoVeiculoEntregue     = "VEICULO_ENTREGUE"     // OS -> ENTREGUE
)

var (
	ErrClienteObrigatorio      = errors.New("clienteId e obrigatorio")
	ErrDestinatarioObrigatorio = errors.New("cliente nao possui e-mail cadastrado")
	ErrDestinatarioInvalido    = errors.New("e-mail do cliente e invalido")
	ErrAssuntoObrigatorio      = errors.New("assunto e obrigatorio")
	ErrConteudoObrigatorio     = errors.New("conteudo e obrigatorio")
	ErrEventoInvalido          = errors.New("tipoEvento nao reconhecido")
	ErrAgregadoObrigatorio     = errors.New("agregado e agregadoId sao obrigatorios")
	ErrJaEnviada               = errors.New("notificacao ja foi enviada")
)

// Origem liga a notificacao ao que a gerou, sem criar uma FK por contexto.
type Origem struct {
	Agregado string
	ID       string
}

type Notificacao struct {
	ID           string
	ClienteID    string
	Canal        string
	TipoEvento   string
	Origem       Origem
	Destinatario string
	Assunto      string
	Conteudo     string
	// ConteudoHTML e opcional: quando presente, o e-mail vai em multipart e o cliente
	// escolhe a melhor versao. Vazio significa apenas texto.
	ConteudoHTML string
	Status       string
	Tentativas   int
	UltimoErro   *string
	CriadaEm     time.Time
	EnviadaEm    *time.Time
}

// Nova monta uma notificacao pronta para a fila, ja validada. Nasce PENDENTE: quem
// dispara so registra a intencao, o envio acontece depois.
func Nova(clienteID, destinatario, tipoEvento, assunto, conteudo string, origem Origem) (Notificacao, error) {
	return NovaComHTML(clienteID, destinatario, tipoEvento, assunto, conteudo, "", origem)
}

// NovaComHTML monta a notificacao com corpo alternativo em HTML.
func NovaComHTML(clienteID, destinatario, tipoEvento, assunto, conteudo, conteudoHTML string, origem Origem) (Notificacao, error) {
	nova := Notificacao{
		ClienteID:    strings.TrimSpace(clienteID),
		Canal:        CanalEmail,
		TipoEvento:   strings.ToUpper(strings.TrimSpace(tipoEvento)),
		Origem:       Origem{Agregado: strings.TrimSpace(origem.Agregado), ID: strings.TrimSpace(origem.ID)},
		Destinatario: strings.TrimSpace(destinatario),
		Assunto:      strings.TrimSpace(assunto),
		Conteudo:     strings.TrimSpace(conteudo),
		ConteudoHTML: strings.TrimSpace(conteudoHTML),
		Status:       StatusPendente,
	}

	if !validation.IsUUID(nova.ClienteID) {
		return Notificacao{}, ErrClienteObrigatorio
	}
	if nova.Destinatario == "" {
		return Notificacao{}, ErrDestinatarioObrigatorio
	}
	if !validation.IsEmail(nova.Destinatario) {
		return Notificacao{}, ErrDestinatarioInvalido
	}
	if !EventoConhecido(nova.TipoEvento) {
		return Notificacao{}, ErrEventoInvalido
	}
	if nova.Origem.Agregado == "" || !validation.IsUUID(nova.Origem.ID) {
		return Notificacao{}, ErrAgregadoObrigatorio
	}
	if nova.Assunto == "" {
		return Notificacao{}, ErrAssuntoObrigatorio
	}
	if nova.Conteudo == "" {
		return Notificacao{}, ErrConteudoObrigatorio
	}
	return nova, nil
}

func EventoConhecido(tipoEvento string) bool {
	switch tipoEvento {
	case EventoDiagnosticoIniciado, EventoOrcamentoPronto, EventoOrcamentoAprovado,
		EventoAguardandoRecursos, EventoRecursosDisponiveis, EventoExecucaoIniciada,
		EventoServicoCancelado, EventoServicoFinalizado, EventoVeiculoEntregue:
		return true
	default:
		return false
	}
}

// MarcarEnviada registra o sucesso. Reenviar algo ja entregue duplicaria o aviso ao
// cliente, entao e recusado.
func (notificacao Notificacao) MarcarEnviada(momento time.Time) (Notificacao, error) {
	if notificacao.Status == StatusEnviada {
		return Notificacao{}, ErrJaEnviada
	}
	notificacao.Status = StatusEnviada
	notificacao.Tentativas++
	notificacao.EnviadaEm = &momento
	notificacao.UltimoErro = nil
	return notificacao, nil
}

// MarcarFalha guarda o motivo para diagnostico e mantem a notificacao reenviavel.
func (notificacao Notificacao) MarcarFalha(motivo string) Notificacao {
	notificacao.Status = StatusFalhou
	notificacao.Tentativas++
	notificacao.UltimoErro = &motivo
	return notificacao
}

// Reenviavel diz se vale tentar de novo: so o que ainda nao chegou ao cliente.
func (notificacao Notificacao) Reenviavel() bool {
	return notificacao.Status == StatusPendente || notificacao.Status == StatusFalhou
}
